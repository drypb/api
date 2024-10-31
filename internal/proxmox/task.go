package proxmox

import (
	"context"

	proxmoxAPI "github.com/luthermonson/go-proxmox"
)

// Task is a wrapper around the proxmoxAPI.Task struct.
type Task struct {
	task *proxmoxAPI.Task
}

// WaitForCompleteStatus is a method that waits for a Proxmox task to reach a
// complete status. It receives the number of times to check the task status
// and the interval between the checks. Returns the status of the task, wheter
// it completed and an error, if any.
func (t *Task) WaitForCompleteStatus(ctx context.Context, timesNum int, steps ...int) (status bool, completed bool, err error) {
	return t.task.WaitForCompleteStatus(ctx, timesNum, steps...)
}
