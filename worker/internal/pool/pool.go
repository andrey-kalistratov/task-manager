package pool

import (
	"log/slog"
	"sync"
	"time"

	"github.com/andrey-kalistratov/task-manager/worker/internal/config"
	"github.com/andrey-kalistratov/task-manager/worker/internal/queue"
	"github.com/andrey-kalistratov/task-manager/worker/internal/task"
)

type Pool struct {
	c  *config.Config
	wg sync.WaitGroup
}

func New(c *config.Config) *Pool {
	return &Pool{c: c, wg: sync.WaitGroup{}}
}

func (p *Pool) Run(c queue.Consumer) error {
	jobs := make(chan task.Task, p.c.WorkerCount)
	slog.Info("pool started working")

	// consumer filling jobs
	errCh := make(chan error, 1)
	go func() {
		errCh <- c.Consume(jobs)
	}()

	for range p.c.WorkerCount {
		p.wg.Add(1)
		go p.worker(jobs)
	}

	select {
	case err := <-errCh:
		return err
	case <-time.After(6 * time.Second):
		slog.Info("Time is up, stop working")
	}

	close(jobs)
	p.wg.Wait()
	slog.Info("pool stopped")
	return nil
}

func (p *Pool) worker(jobs <-chan task.Task) error {
	defer p.wg.Done()
	slog.Info("worker started")

	for t := range jobs {
		if err := t.Do(); err != nil {
			return err
		}
	}

	slog.Info("worker stopped")
	return nil
}
