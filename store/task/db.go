package taskstore

import (
	"github.com/fox-one/pkg/store/db"
	"github.com/yiplee/airdrop/pkg/task"
)

func init() {
	db.RegisterMigrate(func(db *db.DB) error {
		tx := db.Update().Model(task.Task{})

		if err := tx.AutoMigrate(task.Task{}).Error; err != nil {
			return err
		}

		if err := tx.AddUniqueIndex("idx_tasks_trace_id", "trace_id").Error; err != nil {
			return err
		}

		return nil
	})
}
