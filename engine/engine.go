package engine

import (
	"context"
	"time"

	sdk "github.com/fox-one/fox-wallet-sdk"
	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/number"
	"github.com/fox-one/pkg/property"
	"github.com/fox-one/pkg/uuid"
	"github.com/yiplee/airdrop/pkg/task"
)

const checkpointKey = "airdrop_poll_snapshots_checkpoint"

type Engine struct {
	Broker     *sdk.Broker
	UserID     string
	Pin        string
	Tasks      task.Store
	Properties property.Store
}

func (e Engine) Run(ctx context.Context) {
	log := logger.FromContext(ctx).WithField("ctx", "engine")
	ctx = logger.WithContext(ctx, log)

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
			_ = e.run(ctx)
		}
	}
}

func (e Engine) run(ctx context.Context) error {
	log := logger.FromContext(ctx)

	v, err := e.Properties.Get(ctx, checkpointKey)
	if err != nil {
		log.WithError(err).Error("read checkpoint")
		return err
	}

	snapshots, cursor, err := e.Broker.FetchSnapshots(ctx, e.UserID, "", v.String(), "ASC", 500)
	if err != nil {
		log.WithError(err).Error("poll snapshots")
		return err
	}

	for _, snapshot := range snapshots {
		amount := number.Decimal(snapshot.Amount)
		if amount.IsNegative() {
			continue
		}

		t, err := e.Tasks.Find(ctx, snapshot.TraceID)
		if err != nil {
			log.WithError(err).Errorf("find task %s", snapshot.TraceID)
			return err
		}

		if t == nil || snapshot.AssetID != t.AssetID || !amount.Equal(t.Amount) {
			if err := e.refund(ctx, snapshot); err != nil {
				return err
			}

			continue
		}

		t.BrokerID = snapshot.UserID
		t.Payer = snapshot.OpponentID
		payedAt := time.Unix(0, snapshot.CreatedAt)
		t.PayedAt = &payedAt

		if err := e.Tasks.Update(ctx, t); err != nil {
			log.WithError(err).Error("update task")
			return err
		}

		if err := e.handleTask(ctx, t); err != nil {
			return err
		}
	}

	if cursor != v.String() {
		return e.Properties.Save(ctx, checkpointKey, cursor)
	}

	return nil
}

func (e Engine) handleTask(ctx context.Context, t *task.Task) error {
	log := logger.FromContext(ctx).WithField("task", t.ID)
	log.Debug("handle task")

	for _, target := range t.Targets {
		req := &sdk.TransferInput{
			AssetID:    t.AssetID,
			OpponentID: target.UserID,
			Amount:     target.Amount.String(),
			TraceID:    task.UniqueTargetTraceID(t.TraceID, target.UserID),
			Memo:       target.Memo,
		}

		if req.Memo == "" {
			req.Memo = t.Memo
		}

		_, err := e.Broker.Transfer(ctx, e.UserID, e.Pin, req)
		if err != nil {
			log.WithError(err).Errorf("transfer to %s failed", req.OpponentID)

			switch err {
			case sdk.ErrInvalidTrace:
				continue
			case sdk.ErrInvalidRequest:
				continue
			default:
				return err
			}
		}
	}

	log.Info("finish task")
	return nil
}

func (e Engine) refund(ctx context.Context, snapshot *sdk.Snapshot) error {
	log := logger.FromContext(ctx)

	_, err := e.Broker.Transfer(ctx, e.UserID, e.Pin, &sdk.TransferInput{
		AssetID:    snapshot.AssetID,
		OpponentID: snapshot.OpponentID,
		Amount:     snapshot.Amount,
		TraceID:    uuid.Modify(snapshot.SnapshotID, snapshot.TraceID),
		Memo:       "refund by airdrop",
	})

	if err != nil {
		log.WithError(err).Errorf("refund snapshot %s", snapshot.SnapshotID)
	}

	return err
}
