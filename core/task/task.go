package task

import (
	"errors"
	"fmt"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/fox-one/pkg/uuid"
	"github.com/shopspring/decimal"
)

var (
	errInvalidTargetAmount = errors.New("invalid target amount")
)

type (
	Task struct {
		ID        int64           `gorm:"PRIMARY_KEY" json:"id,omitempty"`
		CreatedAt time.Time       `json:"created_at,omitempty"`
		TraceID   string          `gorm:"size:36" json:"trace_id,omitempty"`
		BrokerID  string          `gorm:"size:36" json:"broker_id,omitempty"`
		Payer     string          `gorm:"size:36" json:"payer,omitempty"`
		PayedAt   *time.Time      `json:"payed_at,omitempty"`
		AssetID   string          `gorm:"size:36" json:"asset_id,omitempty"`
		Amount    decimal.Decimal `gorm:"type:DECIMAL(64,8)" json:"amount,omitempty"`
		Memo      string          `gorm:"size:140" json:"memo,omitempty"`
		Targets   Targets         `gorm:"type:TEXT" json:"targets,omitempty"`
	}

	Target struct {
		UserID string          `json:"user_id,omitempty" msgpack:"u,omitempty" valid:"uuid,required"`
		Memo   string          `json:"memo,omitempty" msgpack:"m,omitempty" valid:"stringlength(0|140)"`
		Amount decimal.Decimal `json:"amount,omitempty" msgpack:"a,omitempty"`
	}

	Targets []Target
)

func (t Target) Validate() error {
	if _, err := govalidator.ValidateStruct(t); err != nil {
		return err
	}

	if !t.Amount.IsPositive() {
		return errInvalidTargetAmount
	}

	return nil
}

func (targets Targets) Validate() error {
	users := make(map[string]bool)

	for _, target := range targets {
		if err := target.Validate(); err != nil {
			return fmt.Errorf("target %s is invalid: %w", target.UserID, err)
		}

		if users[target.UserID] {
			return fmt.Errorf("duplicated target with id %s", target.UserID)
		}

		users[target.UserID] = true
	}

	return nil
}

func UniqueTargetTraceID(taskTraceID, targetUserID string) string {
	return uuid.Modify(taskTraceID, targetUserID)
}
