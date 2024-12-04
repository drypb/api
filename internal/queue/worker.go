package queue

import (
	"context"

	"github.com/drypb/api/internal/analysis"
)

type Worker struct {
	job *Job
}

func (w *Worker) work() error {
	a, err := analysis.New(w.job.File, w.job.ID, w.job.Template)
	if err != nil {
		return err
	}
	errRun := a.Run(context.Background())
	if errRun != nil {
		a.Report.Request.Error = errRun.Error()
		a.Report.Save("status")
		a.Report.Save("report")
		errClean := a.Cleanup()
		if errClean != nil {
			return errClean
		}
		return errRun
	}
	return nil
}
