package proxmox

import (
	"context"

	proxmoxAPI "github.com/luthermonson/go-proxmox"
)

// Node is a wrapper around the proxmoxAPI.Node struct.
type Node struct {
	node *proxmoxAPI.Node
}

// VirtualMachine receives a proxmox virtual machine ID and returns a
// VirtualMachine struct and an error, if any.
func (n *Node) VirtualMachine(ctx context.Context, id int) (*VirtualMachine, error) {
	vm, err := n.node.VirtualMachine(ctx, id)
	if err != nil {
		return nil, err
	}
	return &VirtualMachine{virtualMachine: vm}, nil
}
