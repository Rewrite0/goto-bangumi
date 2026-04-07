package taskrunner

import (
	"context"
	"time"

	"goto-bangumi/internal/model"
)

// PhaseResult 阶段执行结果
type PhaseResult struct {
	Err       error         // non-nil 表示任务失败
	PollAfter time.Duration // >0 表示延迟后重新执行当前阶段
}

// PhaseFunc 阶段处理函数
type PhaseFunc func(ctx context.Context, task *model.Task) PhaseResult
