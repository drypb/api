package queue

import (
	"log"

	"github.com/drypb/api/internal/config"
)

var Analysis Queue

type Queue struct {
	jobs       chan *Job
	numWorkers int
}

func (q *Queue) Init() {
	q.jobs = make(chan *Job, config.Api.Queue.Capacity)
	q.numWorkers = config.Api.Queue.MaxWorkers
	q.initWorkers()
}

func (q *Queue) initWorkers() {
	ch := make(chan error)

	for range q.numWorkers {
		go func() {
			for {
				job := <-q.jobs
				w := &Worker{job: job}
				err := w.work()
				if err != nil {
					ch <- err
				}
			}
		}()
	}

	go func() {
		for err := range ch {
			log.Printf("worker: %s\n", err.Error())
		}
	}()
}

func (q *Queue) Enqueue(job *Job) {
	q.jobs <- job
}

func (q *Queue) Close() {
	close(q.jobs)
}
