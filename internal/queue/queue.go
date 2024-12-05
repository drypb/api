package queue

import (
	"log"
	"sync"

	"github.com/drypb/api/internal/config"
)

type Queue struct {
	jobs       chan *Job
	numWorkers int
}

var Analysis *Queue
var once sync.Once

func GetAnalysisQueue() *Queue {
	if Analysis == nil {
		once.Do(
			func() {
				Analysis = &Queue{}
			})
	}
	return Analysis
}

func (q *Queue) Init() {
	apiConfig := config.GetApiConfig()
	q.jobs = make(chan *Job, apiConfig.Queue.Capacity)
	q.numWorkers = apiConfig.Queue.MaxWorkers
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
