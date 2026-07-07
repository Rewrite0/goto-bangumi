package taskrunner

import (
	"context"
	"time"

	"goto-bangumi/internal/model"
)

// PhaseResult 阶段执行结果。
//
// Handler 自己决定错误是否需要重试：
//   - 需要重试时，内部处理错误并只返回 PollAfter。
//   - 不需要重试时，返回 Err，runner 会把任务标记为失败并移出调度。
type PhaseResult struct {
	Err       error         // non-nil 表示任务失败，优先级高于 PollAfter
	PollAfter time.Duration // >0 表示延迟后重新执行当前阶段
}

// PhaseFunc 阶段处理函数
type PhaseFunc func(ctx context.Context, task *model.Task) PhaseResult
