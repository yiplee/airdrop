package task

import (
	"database/sql/driver"

	"github.com/jmoiron/sqlx/types"
	"github.com/vmihailenco/msgpack/v4"
)

func (targets Targets) Value() (driver.Value, error) {
	b, err := msgpack.Marshal(targets)
	if err != nil {
		return nil, err
	}

	g := types.GzippedText(b)
	return g.Value()
}

func (targets *Targets) Scan(src interface{}) error {
	var g types.GzippedText
	if err := g.Scan(src); err != nil {
		return err
	}

	var newTargets Targets
	if err := msgpack.Unmarshal(g, &newTargets); err != nil {
		return err
	}

	*targets = newTargets
	return nil
}
