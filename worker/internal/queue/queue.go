package queue

import (
	"time"

	"github.com/andrey-kalistratov/task-manager/worker/internal/task"
)

type Consumer interface {
	Consume(chan<- task.Task) error
}

// Simple writter
type Mock struct{}

var _ Consumer = (*Mock)(nil)

func NewMock() *Mock {
	return &Mock{}
}

func (m *Mock) Consume(ch chan<- task.Task) error {
	quit := time.After(3 * time.Second)

	for {
		select {
		case <-quit:
			return nil
		default:
			ch <- &task.SleepTask{}
		}
	}
}
