package taskstore

import (
	"context"
	"testing"
	"time"

	"github.com/fox-one/pkg/number"
	db2 "github.com/fox-one/pkg/store/db"
	"github.com/fox-one/pkg/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/yiplee/airdrop/core/task"
)

func TestTaskStore(t *testing.T) {
	ctx := context.Background()

	db, err := db2.Open(db2.SqliteInMemory())
	if err != nil {
		t.Fatal(err)
	}

	if err := db2.Migrate(db); err != nil {
		t.Fatal(err)
	}

	store := New(db)

	a := &task.Task{
		TraceID: uuid.New(),
		AssetID: uuid.New(),
		Amount:  number.Decimal("100"),
		Memo:    "test",
	}

	a.Targets = append(a.Targets, task.Target{
		UserID: uuid.New(),
		Memo:   "test",
		Amount: number.Decimal("100"),
	})

	t.Run("create task", func(t *testing.T) {
		err := store.Create(ctx, a)
		assert.Nil(t, err)
		assert.NotEmpty(t, a.ID)
	})

	t.Run("query nonexistent task", func(t *testing.T) {
		b, err := store.Find(ctx, uuid.New())
		assert.Nil(t, b)
		assert.Nil(t, err)
	})

	t.Run("update task", func(t *testing.T) {
		now := time.Now()
		a.PayedAt = &now
		a.Payer = uuid.New()
		a.BrokerID = uuid.New()
		err := store.Update(ctx, a)
		assert.Nil(t, err)
	})

	t.Run("query existent task", func(t *testing.T) {
		b, err := store.Find(ctx, a.TraceID)
		assert.Nil(t, err)
		assert.Equal(t, a.ID, b.ID)
		assert.NotEmpty(t, b.BrokerID)
		assert.NotEmpty(t, b.Payer)
		assert.NotNil(t, b.PayedAt)
		assert.Equal(t, len(a.Targets), len(b.Targets))
	})
}
