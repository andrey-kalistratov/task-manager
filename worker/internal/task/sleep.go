package task

import (
	"fmt"
	"log/slog"
	"time"
)

type SleepTask struct {
	t time.Duration
}

func (st *SleepTask) Do() error {
	fmt.Println("Done efficent work")
	time.Sleep(st.t)
	return nil
}

func (st *SleepTask) Decode(b []byte) error {
	dur, err := time.ParseDuration(string(b))
	if err != nil {
		slog.Error("provided wrong duration for SleepTask", "err", err)
		return fmt.Errorf("SleepTask Decode error: %w", err)
	}
	st.t = dur
	return nil
}
