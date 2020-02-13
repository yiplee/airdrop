package task

import (
	"context"
)

type Store interface {
	Create(ctx context.Context, task *Task) error
	Update(ctx context.Context, task *Task) error
	Find(ctx context.Context, traceID string) (*Task, error)
}
