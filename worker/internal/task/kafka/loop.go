package kafka

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/segmentio/kafka-go"
)

type Handler interface {
	Handle(ctx context.Context, msg kafka.Message)
}

type Loop struct {
	reader  *kafka.Reader
	handler Handler
	logger  *slog.Logger
}

func (l *Loop) Run(ctx context.Context) error {
	defer func() {
		if err := l.reader.Close(); err != nil {
			l.logger.Error("failed to close reader", "error", err)
		}
	}()

	for {
		msg, err := l.reader.FetchMessage(ctx)
		if err != nil {
			return fmt.Errorf("read message: %w", err)
		}

		l.handler.Handle(ctx, msg)

		if err = l.reader.CommitMessages(ctx, msg); err != nil {
			l.logger.Error("failed to commit message", "error", err)
		}
	}
}
