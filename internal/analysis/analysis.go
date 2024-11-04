// Package analysis is responsable for the malware execution orchestration.
package analysis

import (
	"context"
	"encoding/json"
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

		fi, _ := remoteFile.Stat()
		if fi.Size() < 1 {
			log.Println(ErrFileEmpty)
			//return ErrFileEmpty
		}

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

	modifiedContent := strings.ReplaceAll(string(content), `\`, `/`)
	modifiedContent += "</Registry>"

	var xmlData any
	err = xml.Unmarshal([]byte(modifiedContent), &xmlData)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(xmlData)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonData, &a.Report.Process.WindowsRegisters)
	if err != nil {
		return err
	}

	return nil
}

func (a *Analysis) parseFS() error {
	path := filepath.Join(config.LogPath, a.Report.Request.ID, "fs.txt")
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	modifiedContent := strings.ReplaceAll(string(content), `\`, `/`)
	modifiedContent += "</FileSystem>"

	var xmlData any
	err = xml.Unmarshal([]byte(modifiedContent), &xmlData)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(xmlData)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonData, &a.Report.Process.WindowsFS)
	if err != nil {
		return err
	}

	return nil
}

func (a *Analysis) parseLoad() error {
	path := filepath.Join(config.LogPath, a.Report.Request.ID, "load.txt")
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	modifiedContent := strings.ReplaceAll(string(content), `\`, `/`)
	modifiedContent += "</LoadImage>"

	var xmlData any
	err = xml.Unmarshal([]byte(modifiedContent), &xmlData)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(xmlData)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonData, &a.Report.Process.WindowsBinariesLoaded)
	if err != nil {
		return err
	}

	return nil
}

func (a *Analysis) parseProc() error {
	path := filepath.Join(config.LogPath, a.Report.Request.ID, "proc.txt")
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	modifiedContent := strings.ReplaceAll(string(content), `\`, `/`)
	modifiedContent += "</Process>"

	var xmlData any
	err = xml.Unmarshal([]byte(modifiedContent), &xmlData)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(xmlData)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonData, &a.Report.Process.WindowsProcess)
	if err != nil {
		return err
	}

	return nil
}

// Cleanup deletes the environment after the analysis finished.
func (a *Analysis) Cleanup() error {
	a.env.sshClient.Close()
	err := a.env.destroy()
	if err != nil {
		return fmt.Errorf("failed to destroy environment: %v", err)
	}
	return nil
}
