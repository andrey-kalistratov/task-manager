package task

import "fmt"

type EchoTask struct {
	s string
}

func (et *EchoTask) Decode(b []byte) error {
	et.s = string(b)
	return nil
}

func (et *EchoTask) Do() error {
	fmt.Println(et.s)
	return nil
}
