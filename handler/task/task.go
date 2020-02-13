package taskhandler

import (
	"context"
	"fmt"

	"github.com/asaskevich/govalidator"
	"github.com/fox-one/pkg/logger"
	"github.com/fox-one/pkg/number"
	"github.com/twitchtv/twirp"
	"github.com/yiplee/airdrop/core/task"
	"github.com/yiplee/airdrop/handler/pb"
	"github.com/yiplee/airdrop/handler/views"
)

func New(
	tasks task.Store,
	targetLimit int, // target max count
	brokerID string,
) pb.TaskService {
	return &taskHandler{
		tasks:       tasks,
		targetLimit: targetLimit,
		brokerID:    brokerID,
	}
}

type taskHandler struct {
	tasks       task.Store
	targetLimit int
	brokerID    string
}

func (handler *taskHandler) generateTaskFromReq(req *pb.CreateTaskReq) (*task.Task, error) {
	t := &task.Task{
		TraceID: req.TraceId,
		AssetID: req.AssetId,
		Memo:    req.Memo,
	}

	if !govalidator.IsUUID(t.TraceID) {
		return nil, twirp.InvalidArgumentError("trace_id", "trace_id must be uuid")
	}

	if !govalidator.IsUUID(t.AssetID) {
		return nil, twirp.InvalidArgumentError("asset_id", "asset_id must be uuid")
	}

	if len(t.Memo) > 140 {
		return nil, twirp.InvalidArgumentError("memo", "length of memo must be less than 140")
	}

	for _, target := range req.Targets {
		t.Targets = append(t.Targets, task.Target{
			UserID: target.UserId,
			Memo:   target.Memo,
			Amount: number.Decimal(target.Amount),
		})
	}

	if len(t.Targets) > handler.targetLimit {
		msg := fmt.Sprintf("count of targets must be less than %d", handler.targetLimit)
		return nil, twirp.InvalidArgumentError("targets", msg)
	}

	if err := t.Targets.Validate(); err != nil {
		return nil, twirp.InvalidArgumentError("targets", err.Error())
	}

	t.BrokerID = handler.brokerID
	return t, nil
}

func (handler taskHandler) Create(ctx context.Context, req *pb.CreateTaskReq) (*pb.Task, error) {
	log := logger.FromContext(ctx)

	t, err := handler.generateTaskFromReq(req)
	if err != nil {
		return nil, err
	}

	if existed, err := handler.Find(ctx, &pb.FindTaskReq{TraceId: t.TraceID}); err == nil {
		return existed, nil
	} else if terr, ok := err.(twirp.Error); !ok || terr.Code() != twirp.NotFound {
		return nil, err
	}

	if err := handler.tasks.Create(ctx, t); err != nil {
		log.WithError(err).Error("rpc: cannot create task")
		return nil, err
	}

	return views.Task(t), nil
}

func (handler taskHandler) Find(ctx context.Context, req *pb.FindTaskReq) (*pb.Task, error) {
	log := logger.FromContext(ctx)

	t, err := handler.tasks.Find(ctx, req.TraceId)
	if err != nil {
		log.WithError(err).Error("rpc: cannot find task")
		return nil, err
	}

	if t == nil {
		msg := fmt.Sprintf("task with trace id %s not found", req.TraceId)
		return nil, twirp.NotFoundError(msg)
	}

	return views.Task(t), nil
}
