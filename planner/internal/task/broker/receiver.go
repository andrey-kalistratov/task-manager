package broker

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/segmentio/kafka-go"
)

type Handler interface {
	Handle(ctx context.Context, msg kafka.Message)
}

type Receiver struct {
	reader  *kafka.Reader
	handler Handler
	logger  *slog.Logger
	wg      sync.WaitGroup
}

func NewReceiver(reader *kafka.Reader, handler Handler, logger *slog.Logger) *Receiver {
	return &Receiver{
		reader:  reader,
		handler: handler,
		logger:  logger,
	}
}

func (r *Receiver) Receive(ctx context.Context) error {
	for {
		msg, err := r.reader.FetchMessage(ctx)
		switch {
		case ctx.Err() != nil:
			return nil
		case err != nil:
			return fmt.Errorf("read message: %w", err)
		}

		r.wg.Go(func() {
			r.handle(ctx, msg)
		})
	}
}

func (r *Receiver) handle(ctx context.Context, msg kafka.Message) {
	r.handler.Handle(ctx, msg)

	if err := r.reader.CommitMessages(ctx, msg); err != nil {
		r.logger.Error("failed to commit message", "error", err)
	}
}

func (r *Receiver) Shutdown(ctx context.Context) error {
	if err := r.reader.Close(); err != nil {
		r.logger.Error("failed to close reader", "error", err)
	}

	done := make(chan struct{})
	go func() {
		r.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
