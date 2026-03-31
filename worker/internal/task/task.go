package task

import "fmt"

type Task interface {
	Do() error
	Decode([]byte) error
}

var kvTask = map[string]func() Task{
	"sleep": func() Task { return &SleepTask{} },
	"echo":  func() Task { return &EchoTask{} },
}

func Decode(key, data []byte) (Task, error) {
	tt := string(key)

	ff := kvTask[tt]

	if ff == nil {
		return nil, fmt.Errorf("unknown task type: %s", tt)
	}

	value := ff()
	err := value.Decode(data)

	if err != nil {
		return nil, err
	}

	return value, nil
}
