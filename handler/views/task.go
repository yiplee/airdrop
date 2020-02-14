package views

import (
	"github.com/yiplee/airdrop/handler/pb"
	"github.com/yiplee/airdrop/pkg/task"
)

func Target(t task.Target) *pb.Target {
	return &pb.Target{
		UserId: t.UserID,
		Amount: t.Amount.String(),
		Memo:   t.Memo,
	}
}

func Targets(targets task.Targets) []*pb.Target {
	views := make([]*pb.Target, 0, len(targets))
	for _, t := range targets {
		views = append(views, Target(t))
	}

	return views
}

func Task(t *task.Task) *pb.Task {
	view := &pb.Task{
		TraceId:   t.TraceID,
		CreatedTs: t.CreatedAt.Unix(),
		BrokerId:  t.BrokerID,
		Payer:     t.Payer,
		AssetId:   t.AssetID,
		Amount:    t.Amount.String(),
		Memo:      t.Memo,
		Targets:   Targets(t.Targets),
	}

	if t.PayedAt != nil {
		view.PayedTs = t.PayedAt.Unix()
	}

	return view
}
