package analysis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/drypb/api/internal/data"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

var (
	ErrFileEmpty = errors.New("log file is empty")
)

const (
	analysisTimeout        = 5 * time.Minute
	sampleExecutionTimeout = 2 * time.Minute

	remoteRegPath  = "/Users/administrator/reg.json"
	remoteFSPath   = "/Users/administrator/fs.json"
	remoteLoadPath = "/Users/administrator/load.json"
	remoteProcPath = "/Users/administrator/proc.json"
)

type Analysis struct {
	Report *Report
	env    *Environment
}

// New returns a new analysis object.
func New(header *multipart.FileHeader, id string, template int) (*Analysis, error) {
	filename := header.Filename
	ext := filepath.Ext(filename)
	samplePath := filepath.Join(data.DefaultSamplePath, id+ext)
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

// Run orchestrates the entire analysis proccess.
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
			if err := a.runSample(); err != nil {
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

	err = a.getLog()
	if err != nil {
		return err
	}
	a.Report.LogThis("Results retrieved")

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
	localPath := filepath.Join(data.DefaultSamplePath, a.Report.Request.ID+a.Report.Request.File.Extension)
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
		if err = session.Run(cmd); err != nil {
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

// GetLog remotely gets the malware artifact from the virtual environment.
func (a *Analysis) getLog() error {
	client, err := sftp.NewClient(a.env.sshClient)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %v", err)
	}
	defer client.Close()

	remote := []string{
		remoteRegPath,
		remoteFSPath,
		remoteLoadPath,
		remoteProcPath,
	}
	local := []string{
		"/tmp/reg.json",
		"/tmp/fs.json",
		"/tmp/load.json",
		"/tmp/proc.json",
	}
	for i := 0; i < len(remote); i++ {
		err = a.getLogByName(client, local[i], remote[i])
		if err != nil && err != ErrFileEmpty {
			return err
		}
	}

	// Write report.
	js, err := json.Marshal(a.Report)
	if err != nil {
		return err
	}
	path := data.DefaultReportPath + "/" + a.Report.Request.ID + ".json"
	if err = os.WriteFile(path, js, 0666); err != nil {
		return err
	}

	return nil
}

func (a *Analysis) getLogByName(client *sftp.Client, local string, remote string) error {
	remoteFile, err := client.Open(remote)
	if err != nil {
		return fmt.Errorf("failed to open remote file: %v", err)
	}
	defer remoteFile.Close()

	if fi, _ := remoteFile.Stat(); fi.Size() < 1 {
		return ErrFileEmpty
	}

	localFile, err := os.Create(local)
	if err != nil {
		return fmt.Errorf("failed to create local file: %v", err)
	}
	defer localFile.Close()

	if _, err = remoteFile.WriteTo(localFile); err != nil {
		return fmt.Errorf("failed to copy remote file to local: %v", err)
	}

	if err = a.parseToJSON(local); err != nil {
		return err
	}

	return nil
}

func (a *Analysis) parseToJSON(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	if content[len(content)-3] != ']' {
		// Substitute ',' by ']'.
		content[len(content)-3] = ']'
	}

	// Change from '\' to '/'.
	modifiedContent := strings.ReplaceAll(string(content), "\\", "/")

	switch filename {
	case "/tmp/reg.json":
		if err = json.Unmarshal([]byte(modifiedContent), &a.Report.Process.Reg); err != nil {
			return err
		}
	case "/tmp/fs.json":
		if err = json.Unmarshal([]byte(modifiedContent), &a.Report.Process.FS); err != nil {
			return err
		}
	case "/tmp/load.json":
		if err = json.Unmarshal([]byte(modifiedContent), &a.Report.Process.Load); err != nil {
			return err
		}
	case "/tmp/proc.json":
		if err = json.Unmarshal([]byte(modifiedContent), &a.Report.Process.Proc); err != nil {
			return err
		}
	default:
		return errors.New("not a valid filename")
	}

	return nil
}

// Cleanup deletes the environment and the malware sample
func (a *Analysis) Cleanup() error {
	a.env.sshClient.Close()
	err := a.env.destroy()
	if err != nil {
		return fmt.Errorf("failed to destroy environment: %v", err)
	}
	return nil
}
