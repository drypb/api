package proxmox

import (
	"context"

	proxmoxAPI "github.com/luthermonson/go-proxmox"
)

// Task is a wrapper around [proxmoxAPI.Task].
type Task struct {
	task *proxmoxAPI.Task
}

// WaitForCompleteStatus waits for a Proxmox task to reach a completed status.
// It receives the number of times to check the task status and the interval
// between the checks. Returns the status of the task and wheter it completed
// or not.
func (t *Task) WaitForCompleteStatus(ctx context.Context, timesNum int, steps ...int) (status bool, completed bool, err error) {
	return t.task.WaitForCompleteStatus(ctx, timesNum, steps...)
}
