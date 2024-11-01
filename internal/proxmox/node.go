package proxmox

import (
	"context"

	proxmoxAPI "github.com/luthermonson/go-proxmox"
)

// Node is a wrapper around [proxmoxAPI.Node].
type Node struct {
	node *proxmoxAPI.Node
}

// VirtualMachine returns a [VirtualMachine] by searching a matching id.
func (n *Node) VirtualMachine(ctx context.Context, id int) (*VirtualMachine, error) {
	vm, err := n.node.VirtualMachine(ctx, id)
	if err != nil {
		return nil, err
	}
	return &VirtualMachine{virtualMachine: vm}, nil
}
