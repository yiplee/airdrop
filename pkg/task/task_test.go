package task

import (
	"encoding/json"
	"testing"

	"github.com/fox-one/pkg/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/vmihailenco/msgpack/v4"
)

func TestSqlEncodeTargets(t *testing.T) {
	targets := make(Targets, 1000)
	for idx := range targets {
		targets[idx].UserID = uuid.New()
		targets[idx].Amount = decimal.NewFromInt(int64(idx))
		targets[idx].Memo = uuid.New()
	}

	for _, target := range targets {
		assert.NotEmpty(t, target.UserID)
	}

	v, err := targets.Value()
	assert.Nil(t, err)

	t.Log("msgpack + gzip:", len(v.([]byte)))

	m, _ := msgpack.Marshal(targets)
	t.Log("msgpack:", len(m))

	j, _ := json.Marshal(targets)
	t.Log("json:", len(j))

	t.Run("scan targets", func(t *testing.T) {
		var scanTargets Targets
		err := scanTargets.Scan(v)
		assert.Nil(t, err)
		assert.Len(t, scanTargets, len(targets))

		for idx, target := range scanTargets {
			assert.Equal(t, int64(idx), target.Amount.IntPart())
		}
	})
}
