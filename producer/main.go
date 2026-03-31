package main

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/segmentio/kafka-go"
)

func produce(w *kafka.Writer, key, value []byte) {
	msg := kafka.Message{
		Key:   key,
		Value: value,
	}
	if err := w.WriteMessages(context.Background(), msg); err != nil {
		fmt.Println(err)
	}
}

func fill(w *kafka.Writer, count int) {

	for range count {
		switch rand.Int() % 3 {
		case 0:
			produce(w, []byte("sleep"), []byte("50ms"))
		case 1:
			produce(w, []byte("echo"), []byte("hello world"))
		case 2:
			produce(w, []byte("echo"), []byte("goodbye"))
		}
	}
}

func main() {
	w := kafka.NewWriter(
		kafka.WriterConfig{
			Brokers: []string{"localhost:9091"},
			Topic:   "tasks",
		},
	)

	defer w.Close()

	fill(w, 200)
}
