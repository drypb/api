package proxmox

import (
	"context"

	proxmoxAPI "github.com/luthermonson/go-proxmox"
)

// VirtualMachine is a wrapper around [proxmoxAPI.VirtualMachine].
type VirtualMachine struct {
	virtualMachine *proxmoxAPI.VirtualMachine
}

// Clone clones a [VirtualMachine] and returns it's id.
func (v *VirtualMachine) Clone(ctx context.Context) (int, *Task, error) {
	newID, task, err := v.virtualMachine.Clone(ctx, &proxmoxAPI.VirtualMachineCloneOptions{})
	return newID, &Task{task: task}, err
}

// Start powers on the [VirtualMachine].
func (v *VirtualMachine) Start(ctx context.Context) (*Task, error) {
	t, err := v.virtualMachine.Start(ctx)
	return &Task{task: t}, err
}

// Stop stops the [VirtualMachine].
func (v *VirtualMachine) Stop(ctx context.Context) (*Task, error) {
	t, err := v.virtualMachine.Stop(ctx)
	return &Task{task: t}, err
}

// Delete deletes the [VirtualMachine].
func (v *VirtualMachine) Delete(ctx context.Context) (*Task, error) {
	t, err := v.virtualMachine.Delete(ctx)
	return &Task{task: t}, err
}

// NetworkInterfaces returns the network interfaces information associated with
// the [VirtualMachine].
func (v *VirtualMachine) NetworkInterfaces(ctx context.Context) ([]*NetworkInterface, error) {
	apiIfaces, err := v.virtualMachine.AgentGetNetworkIFaces(ctx)
	if err != nil {
		return nil, err
	}
	var ifaces []*NetworkInterface
	for _, apiIface := range apiIfaces {
		ifaces = append(ifaces, fromAPIInterface(apiIface))
	}
	return ifaces, nil
}

// WaitForAgent waits for the Proxmox QEMU agent to become available on the
// [VirtualMachine].
func (v *VirtualMachine) WaitForAgent(ctx context.Context, seconds int) error {
	return v.virtualMachine.WaitForAgent(ctx, seconds)
}

// AgentExec executes a command on the [VirtualMachine] via QEMU agent.
func (v *VirtualMachine) AgentExec(ctx context.Context, command []string, inputData string) (pid int, err error) {
	return v.virtualMachine.AgentExec(ctx, command, inputData)
}

// Ping pings the [VirtualMachine].
func (v *VirtualMachine) Ping(ctx context.Context) error {
	return v.virtualMachine.Ping(ctx)
}
