package task

import (
	"fmt"
	"time"
)

type Task interface {
	Do() error
}

type SleepTask struct {
	t time.Duration
}

func (st *SleepTask) Do() error {
	fmt.Println("Done efficent work")
	time.Sleep(100 * time.Millisecond)
	return nil
}
