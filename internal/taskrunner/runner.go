package taskrunner

import (
	"context"
	"log/slog"
	"sort"
	"sync"
	"time"

	"goto-bangumi/internal/model"
)

// 要实现的任务有
// 1. task 的阶段转换
// 2. 每个阶段的处理器注册和执行
// 3. 任务完成/失败的处理
// 4. 任务的单独取消

// TaskRunner 任务执行器
type TaskRunner struct {
	store        *TaskStore
	handlers     map[model.TaskPhase][]handlerEntry
	pollInterval time.Duration

	wg sync.WaitGroup
}

// Config TaskRunner 配置
type Config struct {
	PollInterval time.Duration
}

// DefaultConfig 默认配置
func DefaultConfig() Config {
	return Config{
		PollInterval: 5 * time.Second,
	}
}

// NewTaskRunner 创建任务执行器
func NewTaskRunner(config Config) *TaskRunner {
	return &TaskRunner{
		store:        NewTaskStore(),
		handlers:     make(map[model.TaskPhase][]handlerEntry),
		pollInterval: config.PollInterval,
	}
}

// Register 注册阶段处理器
func (r *TaskRunner) Register(phase model.TaskPhase, priority int, handler PhaseHandler) {
	r.handlers[phase] = append(r.handlers[phase], handlerEntry{
		priority: priority,
		handler:  handler,
	})
	// 按优先级排序
	sort.Slice(r.handlers[phase], func(i, j int) bool {
		return r.handlers[phase][i].priority < r.handlers[phase][j].priority
	})
}

// Store 返回任务存储
func (r *TaskRunner) Store() *TaskStore {
	return r.store
}

func (r *TaskRunner) Add(ctx context.Context, task *model.Task) bool {
	if added := r.store.Add(task); added {
		go r.executeTask(ctx, task)
		return true
	}
	return false
}

// executeTask 执行单个任务
func (r *TaskRunner) executeTask(ctx context.Context, task *model.Task) {
	ctx, cancel := context.WithCancel(ctx)
	task.Cancel = cancel
	// 每次执行都只是简单的把当前阶段的处理器跑一遍
	// 然后下一次再跑一次
	// 每次都会把 HandlerIndex ++, 以便下次从下一个处理器开始
	// 如果本身把 HandlerIndex-- 了, 那么就会重复跑这个处理器
	// 对于跑的结果, 如果返回了 NextPhase, 那么就直接跳到下一个阶段, 跳过剩下的处理器
	// 对于失败要分情况, 有的错误可接受, 这时候就继续跑下一个处理器
	// 有的错误不可接受, 这时候就直接把任务标记为失败
	// 目前每个阶段有这么几个处理器:
	// PhaseChecking:
	//   1. 检查下载是否成功添加 (不可接受错误)
	// PhaseDownloading:
	//   1. 检查下载状态 (可接受错误, 比如还在下载中)
	// PhaseRenaming:
	//   1. 重命名文件 (不可接受错误)
	// Complete:
	//   1. 用以通知
	for {
		handlers := r.handlers[task.Phase]
		// 如果当前阶段没有处理器或到了最后一个 Handler，直接前往下一个阶段
		if task.HandlerIndex >= len(handlers) {
			// nextPhase 返回 false, 表明到了终态，结束任务
			if res := r.nextPhase(task); !res {
				return
			}
			continue
		}
		entry := handlers[task.HandlerIndex]
		result := entry.handler.Handle(ctx, task)
		r.applyResult(task, result)
	}
}

// applyResult 应用执行结果
func (r *TaskRunner) applyResult(task *model.Task, result HandlerResult) {
	//TODO: 错误分为可接受和不可接受两种
	// 如果是可接受错误, 那么就继续执行下一个处理器
	// 如果是不可接受错误, 那么就把任务标记为失败
	if result.Error != nil {
		task.ErrorMsg = result.Error.Error()
		task.Advance(model.PhaseFailed)
		slog.Error("[task runner] 任务失败",
			"torrent", task.Torrent.Name,
			"phase", task.Phase,
			"error", result.Error)
		return
	}
	// 更新阶段
	if result.NextPhase != 0 {
		task.Advance(result.NextPhase)
	}
	task.HandlerIndex++
}

// nextPhase 前往下一个阶段
func (r *TaskRunner) nextPhase(task *model.Task) bool {
	oldPhase := task.Phase
	task.Advance()
	if task.Phase.IsTerminal() {
		// 将任务删除
		task.Cancel()
		r.store.Remove(task.Torrent.Link)
		slog.Info("[task runner] 任务完成",
			"torrent", task.Torrent.Name,
			"final_phase", task.Phase)
		return false
	} else {
		slog.Debug("[task runner] 阶段变更",
			"torrent", task.Torrent.Name,
			"from", oldPhase,
			"to", task.Phase)
	}
	return true
}

func (r *TaskRunner) Cancel(link string) {
	// 调用的时候无法保证任务有cancel
	// 所以应该把 Task设为 End 状态
	task := r.store.Get(link)
	task.Advance(model.PhaseEnd)
	if task.Cancel != nil {
		task.Cancel()
	}
}
