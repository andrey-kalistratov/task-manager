package queue

import (
	"context"
	"log/slog"

	"github.com/andrey-kalistratov/task-manager/worker/internal/task"
	k "github.com/segmentio/kafka-go"
)

type Kafka struct {
	r *k.Reader
}

var _ Consumer = (*Kafka)(nil)

func NewKafka(address []string, topic, groupID string) *Kafka {
	return &Kafka{r: k.NewReader(
		k.ReaderConfig{
			Brokers: address,
			Topic:   topic,
			GroupID: groupID,
		},
	),
	}
}

func (k *Kafka) Consume(ctx context.Context, ch chan<- task.Task) error {
	defer k.r.Close()

	for {
		msg, err := k.r.FetchMessage(ctx)

		if err != nil {
			slog.Error("unable to FetchMessage", "err", err)
			return err
		}

		t, err := task.Decode(msg.Key, msg.Value)

		if err != nil {
			slog.Error("provided wrong task", "err", err)
			if err := k.r.CommitMessages(ctx, msg); err != nil {
				slog.Error("unable to CommitMessages", "err", err)
				return err
			}
			continue
		}

		ch <- t
		if err := k.r.CommitMessages(ctx, msg); err != nil {
			slog.Error("unable to CommitMessages", "err", err)
			return err
		}
	}
}
