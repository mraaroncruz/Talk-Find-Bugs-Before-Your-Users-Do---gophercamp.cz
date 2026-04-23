package runs

import (
	"context"
	"log/slog"
	"sync"
)

type Job struct {
	RunID    int64
	RunnerID int64
	Tiles    [][2]int
}

type WorkerPool struct {
	svc     *Service
	jobs    chan Job
	workers int
	wg      sync.WaitGroup
	stop    chan struct{}
}

func NewWorkerPool(svc *Service, workers int) *WorkerPool {
	return &WorkerPool{
		svc:     svc,
		jobs:    make(chan Job, 100),
		workers: workers,
		stop:    make(chan struct{}),
	}
}

func (p *WorkerPool) Start() {
	for i := range p.workers {
		p.wg.Add(1)
		go p.worker(i)
	}
	slog.Info("worker pool started", "workers", p.workers)
}

func (p *WorkerPool) Stop() {
	close(p.stop)
	p.wg.Wait()
}

func (p *WorkerPool) Submit(job Job) {
	select {
	case p.jobs <- job:
	default:
		slog.Warn("worker pool full, dropping job", "run_id", job.RunID)
	}
}

func (p *WorkerPool) worker(id int) {
	defer p.wg.Done()
	for {
		select {
		case <-p.stop:
			return
		case job := <-p.jobs:
			ctx := context.Background()
			if err := p.svc.ProcessRun(ctx, job.RunID, job.RunnerID, job.Tiles); err != nil {
				slog.Error("process run failed",
					"worker", id,
					"run_id", job.RunID,
					"error", err,
				)
			}
		}
	}
}
