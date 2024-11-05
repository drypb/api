package analysis

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/luthermonson/go-proxmox"
	"golang.org/x/crypto/ssh"
)

const (
	sshUser                 = "administrator"
	sshKeyPath              = "/run/secrets/key"
	sshMaxAttempts          = 6
	sshDelayAttemptsSeconds = 10

	startMaxAttempts          = 12
	startDelayAttemptsSeconds = 5

	stopMaxAttempts          = 6
	stopDelayAttemptsSeconds = 5

	waitAgentSeconds = 120
)

type Environment struct {
	templateID int

	vm   *proxmox.VirtualMachine
	node *proxmox.Node

	sshClient *ssh.Client
}

// NewProxmoxClient creates a new [proxmox.Client].
func newProxmoxClient() *proxmox.Client {
	url := getFatalEnv("PROXMOX_URL")
	id := getFatalEnv("PROXMOX_TOKEN_ID")
	secret := getFatalEnv("PROXMOX_TOKEN_SECRET")

	// Maybe change to use TLS when in production.
	insecureHTTPClient := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	client := proxmox.NewClient(
		url,
		proxmox.WithHTTPClient(&insecureHTTPClient),
		proxmox.WithAPIToken(id, secret),
	)
	if client == nil {
		log.Fatal("Failed to create proxmox client")
	}

	return client
}

// Create creates VM from template.
func (e *Environment) create() error {
	var err error

	err = e.cloneTemplate()
	if err != nil {
		return err
	}
	err = e.start()
	if err != nil {
		return err
	}
	err = e.vm.WaitForAgent(context.Background(), waitAgentSeconds)
	if err != nil {
		return err
	}
	err = e.connectSSH()
	if err != nil {
		return err
	}

	return nil
}

func (e *Environment) cloneTemplate() error {
	template, err := e.getTemplateVM()
	if err != nil {
		return err
	}
	options := proxmox.VirtualMachineCloneOptions{}
	id, _, err := template.Clone(context.Background(), &options)
	if err != nil {
		return err
	}
	e.vm, err = e.node.VirtualMachine(context.Background(), id)
	if err != nil {
		return err
	}
	return nil
}

func (e *Environment) getTemplateVM() (*proxmox.VirtualMachine, error) {
	client := newProxmoxClient()
	var err error
	e.node, err = client.Node(context.Background(), getFatalEnv("PROXMOX_NODE"))
	if err != nil {
		return nil, err
	}
	template, err := e.node.VirtualMachine(context.Background(), e.templateID)
	if err != nil {
		return nil, err
	}
	return template, nil
}

func (e *Environment) start() error {
	attempts := startMaxAttempts
	delay := startDelayAttemptsSeconds * time.Second
	return e.startWithRetry(attempts, delay)
}

func (e *Environment) startWithRetry(attempts int, delay time.Duration) error {
	task, err := e.vm.Start(context.Background())
	if err != nil {
		return err
	}
	for i := 0; i < attempts; i++ {
		_, completed, _ := task.WaitForCompleteStatus(context.Background(), 10, 5)
		if completed {
			return nil
		}
		time.Sleep(delay)
	}
	return fmt.Errorf("failed to start virtual machine after %d attempts", attempts)
}

// ConnectSSH creates a SSH connection with the guest VM.
func (e *Environment) connectSSH() error {
	key, err := readPrivateKey()
	if err != nil {
		return err
	}
	config := ssh.ClientConfig{
		User: sshUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	err = e.dialSSH(&config)
	if err != nil {
		return err
	}
	return nil
}

// GetIP tries to get valid IPv4 of VM.
func (e *Environment) getIP() (string, error) {
	iFaces, err := e.vm.AgentGetNetworkIFaces(context.Background())
	if err != nil {
		return "", err
	}

	for _, iFace := range iFaces {
		for _, ip := range iFace.IPAddresses {
			if ip.IPAddressType == "ipv4" {
				if ip.IPAddress != "127.0.0.1" && !strings.HasPrefix(ip.IPAddress, "169.254.") {
					return ip.IPAddress, nil
				}
			}
		}
	}

	return "", fmt.Errorf("failed to get IP address")
}

func readPrivateKey() (ssh.Signer, error) {
	path := os.ExpandEnv(sshKeyPath)
	key, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %v", err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}
	return signer, nil
}

func (e *Environment) dialSSH(config *ssh.ClientConfig) error {
	attempts := sshMaxAttempts
	delay := sshDelayAttemptsSeconds * time.Second
	return e.dialSSHWithRetry(config, attempts, delay)
}

func (e *Environment) dialSSHWithRetry(config *ssh.ClientConfig, attempts int, delay time.Duration) (err error) {
	for i := 0; i < attempts; i++ {
		ip, err := e.getIP()
		if err != nil {
			log.Printf("Failed to get IP address on attempt %d/%d: %v", i+1, attempts, err)
			time.Sleep(delay)
			continue
		}
		e.sshClient, err = ssh.Dial("tcp", ip+":22", config)
		if err == nil {
			log.Println("SSH connection established!")
			return nil
		}
		log.Printf("Failed to connect via SSH on attempt %d/%d: %v. Retrying in %v...", i+1, attempts, err, delay)
		time.Sleep(delay)
	}
	log.Printf("Failed to connect via SSH after %d attempts", attempts)
	return errors.New("failed to dial SSH")
}

// Destroy deletes the environment.
func (e *Environment) destroy() error {
	err := e.stop()
	if err != nil {
		return err
	}
	_, err = e.vm.Delete(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (e *Environment) stop() error {
	attempts := stopMaxAttempts
	delay := stopDelayAttemptsSeconds * time.Second
	err := e.stopWithRetry(attempts, delay)
	if err != nil {
		return err
	}
	return nil
}

func (e *Environment) stopWithRetry(attempts int, delay time.Duration) error {
	task, err := e.vm.Stop(context.Background())
	if err != nil {
		return fmt.Errorf("failed to stop virtual machine: %v", err)
	}
	for i := 0; i < attempts; i++ {
		_, completed, _ := task.WaitForCompleteStatus(context.Background(), 10, 2)
		if completed {
			return nil
		}
		time.Sleep(delay)
	}
	return fmt.Errorf("failed to stop virtual machine after %d attempts", attempts)
}
