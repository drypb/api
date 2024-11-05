// Package analysis is responsable for the malware execution orchestration.
package analysis

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/drypb/api/internal/config"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

var (
	ErrFileEmpty = errors.New("log file is empty") // ErrFileEmpty occurs when the driver log file is empty.
)

const (
	analysisTimeout        = 5 * time.Minute
	sampleExecutionTimeout = 2 * time.Minute

	remotePath = "/Users/administrator/"
)

type Analysis struct {
	Report *Report      // Report represents the final artifact of the analysis process.
	env    *Environment // Environment represents where the analysis will occur.
}

// New creates an [Analysis] object.
func New(header *multipart.FileHeader, id string, template int) (*Analysis, error) {
	filename := header.Filename
	ext := filepath.Ext(filename)
	samplePath := filepath.Join(config.SamplePath, id+ext)
	mimeType, err := getMimeType(samplePath)
	if err != nil {
		return nil, err
	}
	size, err := getSize(samplePath)
	if err != nil {
		return nil, err
	}
	lastModified, err := getLastModified(samplePath)
	if err != nil {
		return nil, err
	}
	md5sum, err := getMD5Sum(samplePath)
	if err != nil {
		return nil, err
	}
	sha1sum, err := getSHA1Sum(samplePath)
	if err != nil {
		return nil, err
	}
	sha256sum, err := getSHA256Sum(samplePath)
	if err != nil {
		return nil, err
	}

	a := &Analysis{
		env: &Environment{
			templateID: template,
		},
		Report: &Report{
			Request: RequestMetadata{
				Status:        "Running",
				ID:            id,
				TemplateID:    template,
				DriverVersion: getDriverVersion(template),
				StartTime:     time.Now().Format(TimeFormat),
				File: FileMetadata{
					Filename:     filename,
					Extension:    ext,
					MimeType:     mimeType,
					Size:         size,
					LastModified: lastModified,
					MD5Sum:       md5sum,
					SHA1Sum:      sha1sum,
					SHA256Sum:    sha256sum,
				},
			},
		},
	}

	err = a.Report.Save("status")
	if err != nil {
		return nil, err
	}

	err = a.Report.Save("report")
	if err != nil {
		return nil, err
	}

	return a, nil
}

// Run starts an [Analysis].
func (a *Analysis) Run(parent context.Context) error {
	ctx, cancel := context.WithTimeout(parent, analysisTimeout)
	defer cancel()

	ch := make(chan error)
	go func() {
		ch <- a.runWithoutCtx()
	}()

	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		return fmt.Errorf("Analysis reached timeout (%v)", analysisTimeout)
	}
}
func (a *Analysis) runWithoutCtx() error {
	a.Report.LogThis("Providing environment...")
	err := a.env.create()
	if err != nil {
		return err
	}
	a.Report.LogThis("Analysis environment created")

	err = a.sendSample()
	if err != nil {
		return err
	}
	a.Report.LogThis("Sample sent to environment")

	ch1 := make(chan error)
	ch2 := make(chan error)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				err := a.env.vm.Ping(ctx)
				if err != nil {
					ch1 <- err
					return
				}
			}
		}
	}()

	go func() {
		select {
		case <-ctx.Done():
			return
		default:
			a.Report.LogThis("Analysis started")
			err := a.runSample()
			if err != nil {
				ch2 <- err
				return
			}
			a.Report.LogThis("Analysis finished")
			ch2 <- nil
		}
	}()

	select {
	case err := <-ch1:
		if err != nil {
			cancel()
			return err
		}
	case err := <-ch2:
		cancel()
		if err != nil {
			return err
		}
	}

	err = a.getResults()
	if err != nil {
		return err
	}
	a.env.sshClient.Close()
	a.Report.LogThis("Results retrieved")

	err = a.parseResults()
	if err != nil {
		return err
	}

	a.Report.Request.Status = "Completed"
	err = a.Report.Save("status")
	if err != nil {
		return err
	}

	err = a.Report.Save("report")
	if err != nil {
		return err
	}

	err = a.Cleanup()
	if err != nil {
		return err
	}
	a.Report.LogThis("Environment deleted")

	return nil
}

// SendSample uploads the malware sample to the virtual environment.
func (a *Analysis) sendSample() error {
	localPath := filepath.Join(config.SamplePath, a.Report.Request.ID+a.Report.Request.File.Extension)
	remotePath := "./sample" + a.Report.Request.File.Extension

	sftpClient, err := sftp.NewClient(a.env.sshClient)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %v", err)
	}
	defer sftpClient.Close()

	src, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %v", err)
	}
	defer src.Close()

	dst, err := sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("failed to create remote file: %v", err)
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to copy local file to remote: %v", err)
	}

	return nil
}

// RunSample remotely executes a program that will execute the malware sample
// inside the virtual environment using the amaterasu driver to inspect its
// process.
func (a *Analysis) runSample() error {
	session, err := a.env.sshClient.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer session.Close()

	ctx, cancel := context.WithTimeout(context.Background(), sampleExecutionTimeout)
	defer cancel()

	ch := make(chan error)
	go func() {
		file := "sample" + a.Report.Request.File.Extension
		cmd := "/Users/administrator/amaterasu.client.exe L M 1024 n " + file + " a"
		err := session.Run(cmd)
		if err != nil {
			ch <- fmt.Errorf("failed to start driver: %v", err)
		}
	}()

	select {
	case <-ctx.Done():
		if err = session.Signal(ssh.SIGKILL); err != nil {
			return fmt.Errorf("failed to kill client: %v", err)
		}
	case err = <-ch:
		if err != nil {
			return err
		}
	}

	return nil
}

// GetResults remotely gets the malware artifact from virtual environment.
func (a *Analysis) getResults() error {
	client, err := sftp.NewClient(a.env.sshClient)
	if err != nil {
		return err
	}
	defer client.Close()

	files := []string{
		"reg.txt",
		"fs.txt",
		"load.txt",
		"proc.txt",
	}

	for _, file := range files {
		remoteFile, err := client.Open(filepath.Join(remotePath, file))
		if err != nil {
			return err
		}
		defer remoteFile.Close()

		logIDDir := filepath.Join(config.LogPath, a.Report.Request.ID)
		err = os.MkdirAll(logIDDir, 0750)
		if err != nil {
			return err
		}
		localPath := filepath.Join(logIDDir, file)
		localFile, err := os.Create(localPath)
		if err != nil {
			return err
		}
		defer localFile.Close()

		_, err = remoteFile.WriteTo(localFile)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *Analysis) parseResults() error {
	var err error
	err = a.parseReg()
	if err != nil {
		return err
	}
	err = a.parseFS()
	if err != nil {
		return err
	}
	err = a.parseLoad()
	if err != nil {
		return err
	}
	err = a.parseProc()
	if err != nil {
		return err
	}
	return nil
}

func (a *Analysis) parseReg() error {
	path := filepath.Join(config.LogPath, a.Report.Request.ID, "reg.txt")
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if len(content) == 0 {
		log.Println("reg.txt empty")
		return nil
	}

	modifiedContent := strings.ReplaceAll(string(content), `\`, `/`)
	modifiedContent += "</registry>"

	var Registry struct {
		Log []struct {
			Date              string `xml:"date"`
			Time              string `xml:"time"`
			InfoType          string `xml:"info_type"`
			RegistryOperation string `xml:"registry_operation"`
			Name              string `xml:"name"`
			DataType          string `xml:"data_type"`
			Data              string `xml:"data"`
		} `xml:"log"`
	}
	err = xml.Unmarshal([]byte(modifiedContent), &Registry)
	if err != nil {
		return err
	}

	for _, entry := range Registry.Log {
		register := WindowsRegisters{
			Date:              entry.Date,
			Time:              entry.Time,
			InfoType:          entry.InfoType,
			RegistryOperation: entry.RegistryOperation,
			Name:              entry.Name,
			DataType:          entry.DataType,
			Data:              entry.Data,
		}
		a.Report.Process.WindowsRegisters = append(a.Report.Process.WindowsRegisters, register)
	}

	return nil
}

func (a *Analysis) parseFS() error {
	path := filepath.Join(config.LogPath, a.Report.Request.ID, "fs.txt")
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if len(content) == 0 {
		log.Println("fs.txt empty")
		return nil
	}

	modifiedContent := strings.ReplaceAll(string(content), `\`, `/`)
	modifiedContent += "</file_system>"

	var FileSystem struct {
		Log []struct {
			Date       string `xml:"date"`
			Time       string `xml:"time"`
			InfoType   string `xml:"info_type"`
			MJFunc     string `xml:"mj_func"`
			PID        string `xml:"pid"`
			TID        string `xml:"tid"`
			SID        string `xml:"sid"`
			TokenType  string `xml:"token_type"`
			Privileges struct {
				Privilege []struct {
					Name  string `xml:"name"`
					Value string `xml:"value"`
				} `xml:"privilege"`
			} `xml:"privileges"`
			ElevationStatus string `xml:"elevation_status"`
			ImageName       string `xml:"image_name"`
			Path            string `xml:"path"`
			FileName        string `xml:"file_name"`
		} `xml:"log"`
	}
	err = xml.Unmarshal([]byte(modifiedContent), &FileSystem)
	if err != nil {
		return err
	}

	for _, entry := range FileSystem.Log {
		var privileges []Privilege
		for _, p := range entry.Privileges.Privilege {
			privileges = append(privileges, Privilege{
				Name:  p.Name,
				Value: p.Value,
			})
		}
		fs := WindowsFileSystem{
			Date:            entry.Date,
			Time:            entry.Time,
			InfoType:        entry.InfoType,
			MJFunc:          entry.MJFunc,
			PID:             entry.PID,
			TID:             entry.TID,
			SID:             entry.SID,
			TokenType:       entry.TokenType,
			Privileges:      privileges,
			ElevationStatus: entry.ElevationStatus,
			ImageName:       entry.ImageName,
			Path:            entry.Path,
			FileName:        entry.FileName,
		}
		a.Report.Process.WindowsFS = append(a.Report.Process.WindowsFS, fs)
	}

	return nil
}

func (a *Analysis) parseLoad() error {
	path := filepath.Join(config.LogPath, a.Report.Request.ID, "load.txt")
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if len(content) == 0 {
		log.Println("load.txt empty")
		return nil
	}

	modifiedContent := strings.ReplaceAll(string(content), `\`, `/`)
	modifiedContent += "</load_image>"

	var LoadImage struct {
		Log []struct {
			Date          string `xml:"date"`
			Time          string `xml:"time"`
			InfoType      string `xml:"info_type"`
			PID           string `xml:"pid"`
			FullImageName string `xml:"full_image_name"`
			FileName      string `xml:"file_name"`
		} `xml:"log"`
	}
	err = xml.Unmarshal([]byte(modifiedContent), &LoadImage)
	if err != nil {
		return err
	}

	for _, entry := range LoadImage.Log {
		binaryLoaded := WindowsBinariesLoaded{
			Date:          entry.Date,
			Time:          entry.Time,
			InfoType:      entry.InfoType,
			PID:           entry.PID,
			FullImageName: entry.FullImageName,
			FileName:      entry.FileName,
		}
		a.Report.Process.WindowsBinariesLoaded = append(a.Report.Process.WindowsBinariesLoaded, binaryLoaded)
	}

	return nil
}

func (a *Analysis) parseProc() error {
	path := filepath.Join(config.LogPath, a.Report.Request.ID, "proc.txt")
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if len(content) == 0 {
		log.Println("proc.txt empty")
		return nil
	}

	modifiedContent := strings.ReplaceAll(string(content), `\`, `/`)
	modifiedContent += "</process>"

	var Process struct {
		Log []struct {
			Date       string `xml:"date"`
			Time       string `xml:"time"`
			InfoType   string `xml:"info_type"`
			PPID       string `xml:"ppid"`
			PID        string `xml:"pid"`
			Operation  string `xml:"operation"`
			TokenType  string `xml:"token_type"`
			Privileges struct {
				Privilege []struct {
					Name  string `xml:"name"`
					Value string `xml:"value"`
				} `xml:"privilege"`
			} `xml:"privileges"`
			ElevationStatus string `xml:"elevation_status"`
			ParentName      string `xml:"parent_name"`
			ChildName       string `xml:"child_name"`
		} `xml:"log"`
	}
	err = xml.Unmarshal([]byte(modifiedContent), &a.Report.Process.WindowsProcess)
	if err != nil {
		return err
	}

	for _, entry := range Process.Log {
		var privileges []Privilege
		for _, p := range entry.Privileges.Privilege {
			privileges = append(privileges, Privilege{
				Name:  p.Name,
				Value: p.Value,
			})
		}
		proc := WindowsProcess{
			Date:            entry.Date,
			Time:            entry.Time,
			InfoType:        entry.InfoType,
			PPID:            entry.PPID,
			PID:             entry.PID,
			Operation:       entry.Operation,
			TokenType:       entry.TokenType,
			Privileges:      privileges,
			ElevationStatus: entry.ElevationStatus,
			ParentName:      entry.ParentName,
			ChildName:       entry.ChildName,
		}
		a.Report.Process.WindowsProcess = append(a.Report.Process.WindowsProcess, proc)
	}

	return nil
}

// Cleanup deletes the environment after the analysis finished.
func (a *Analysis) Cleanup() error {
	err := a.env.destroy()
	if err != nil {
		return fmt.Errorf("failed to destroy environment: %v", err)
	}
	return nil
}
