package taskrunner

import (
	"context"
	"time"

	"goto-bangumi/internal/model"
)

// HandlerResult Handler 执行结果
type HandlerResult struct {
	NextPhase     model.TaskPhase // 下一阶段
	ScheduleAfter time.Duration   // 延迟执行（用于轮询）
	Error         error
	ShouldRetry   bool
}

// PhaseHandler 阶段处理器接口
type PhaseHandler interface {
	Handle(ctx context.Context, task *model.Task) HandlerResult
}

// handlerEntry 存储 handler 和优先级
type handlerEntry struct {
	priority int
	handler  PhaseHandler
}
