package taskstore

import (
	"context"

	"github.com/fox-one/pkg/store/db"
	"github.com/yiplee/airdrop/core/task"
)

func New(db *db.DB) task.Store {
	return &taskStore{db: db}
}

type taskStore struct {
	db *db.DB
}

func (t taskStore) Create(ctx context.Context, task *task.Task) error {
	return t.db.Update().Create(task).Error
}

func toUpdateParams(task *task.Task) map[string]interface{} {
	return map[string]interface{}{
		"broker_id": task.BrokerID,
		"payer":     task.Payer,
		"payed_at":  task.PayedAt,
	}
}

func (t taskStore) Update(ctx context.Context, task *task.Task) error {
	return t.db.Update().Model(task).Updates(toUpdateParams(task)).Error
}

func (t taskStore) Find(ctx context.Context, traceID string) (*task.Task, error) {
	var tt task.Task
	if err := t.db.View().Where("trace_id = ?", traceID).First(&tt).Error; err != nil {
		if db.IsErrorNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	return &tt, nil
}
