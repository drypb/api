// The Environment is where the analysis will be running on.
package analysis

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/drypb/api/internal/proxmox"

	"golang.org/x/crypto/ssh"
)

type Environment struct {
	templateID int
	vm         *proxmox.VirtualMachine
	node       *proxmox.Node
	sshClient  *ssh.Client
	ip         string
}

// Create creates an Environment from a template with ID templateID.
func (e *Environment) create() (err error) {
	if err := e.cloneTemplate(); err != nil {
		return err
	}
	if err := e.startVM(); err != nil {
		return err
	}
	if err = e.vm.WaitForAgent(context.Background(), 120); err != nil {
		return err
	}
	if err := e.handleSSH(); err != nil {
		return err
	}
	return nil
}

func (e *Environment) cloneTemplate() error {
	template, err := e.getTemplateVM()
	if err != nil {
		return err
	}
	id, _, err := template.Clone(context.Background())
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
	client := proxmox.NewClient()
	node := os.Getenv("PROXMOX_NODE")
	if node == "" {
		log.Fatal("PROXMOX_NODE is empty")
	}
	var err error
	e.node, err = client.Node(context.Background(), node)
	if err != nil {
		return nil, err
	}
	template, err := e.node.VirtualMachine(context.Background(), e.templateID)
	if err != nil {
		return nil, err
	}
	return template, nil
}

func (e *Environment) startVM() error {
	attempts := 12
	delay := 5 * time.Second
	return e.startVMWithRetry(attempts, delay)
}

func (e *Environment) startVMWithRetry(attempts int, delay time.Duration) error {
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

// HandleSSH creates a SSH connection with the guest VM.
func (e *Environment) handleSSH() error {
	if err := e.getIP(); err != nil {
		return fmt.Errorf("failed to get virtual machine IP: %v", err)
	}
	key, err := readPrivateKey()
	if err != nil {
		return err
	}
	config := createSSHClientConfig(key)
	if err = e.connectSSH(config); err != nil {
		return err
	}
	return nil
}

// GetIP tries to get a valid IPv4 to the environment virtual machine.
func (e *Environment) getIP() error {
	iFaces, err := e.vm.NetworkInterfaces(context.Background())
	if err != nil {
		return fmt.Errorf("failed to network interfaces: %v", err)
	}
	for _, iFace := range iFaces {
		for _, ip := range iFace.IPAddresses {
			if ip.IPAddressType == "ipv4" {
				if ip.IPAddress != "127.0.0.1" && !strings.HasPrefix(ip.IPAddress, "169.254.") {
					e.ip = ip.IPAddress
					return nil
				}
			}
		}
	}
	return err
}

func readPrivateKey() (ssh.Signer, error) {
	path := os.ExpandEnv("/run/secrets/key")
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

func createSSHClientConfig(signer ssh.Signer) *ssh.ClientConfig {
	return &ssh.ClientConfig{
		User: "administrator",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
}

func (e *Environment) connectSSH(config *ssh.ClientConfig) error {
	attempts := 6
	delay := 5 * time.Second
	return e.connectSSHWithRetry(config, attempts, delay)
}

func (e *Environment) connectSSHWithRetry(config *ssh.ClientConfig, attempts int, delay time.Duration) (err error) {
	for i := 0; i < attempts; i++ {
		err := e.getIP()
		if e.ip == "" {
			log.Println("ip empty")
		}
		if err != nil {
			log.Println(err)
		}
		e.sshClient, err = ssh.Dial("tcp", e.ip+":22", config)
		if err == nil {
			log.Println("SSH connection established!")
			break
		}
		log.Println(err)
		log.Printf("Trying again in %v... (%s)\n", delay, e.ip)
		time.Sleep(delay)
	}
	if err != nil || e.sshClient == nil {
		return errors.New("failed to connected")
	}
	return nil
}

// Destroy deletes the environment.
func (e *Environment) destroy() error {
	attempts := 6
	delay := 5 * time.Second
	err := e.stopWithRetry(attempts, delay)
	if err != nil {
		return err
	}
	_, err = e.vm.Delete(context.Background())
	if err != nil {
		return fmt.Errorf("failed to delete virtual machine: %v", err)
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
