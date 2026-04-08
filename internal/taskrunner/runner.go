package taskrunner

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"goto-bangumi/internal/model"
)

// phaseEntry 阶段配置
type phaseEntry struct {
	phase   model.TaskPhase
	handler PhaseFunc
}

// TaskRunner 任务执行器
type TaskRunner struct {
	store     *TaskStore
	phases    []phaseEntry
	queue     chan *model.Task
	addingSem chan struct{} // 流水线槽位信号量
	wg        sync.WaitGroup
	cancel    context.CancelFunc
}

// New 创建任务执行器
func New(queueSize, maxConcurrency int) *TaskRunner {
	if maxConcurrency <= 0 {
		maxConcurrency = 4
	}
	if queueSize <= 0 {
		queueSize = 64
	}
	return &TaskRunner{
		store:     NewTaskStore(),
		queue:     make(chan *model.Task, queueSize),
		addingSem: make(chan struct{}, maxConcurrency),
	}
}

// Register 注册阶段处理器
func (r *TaskRunner) Register(phase model.TaskPhase, handler PhaseFunc) {
	r.phases = append(r.phases, phaseEntry{
		phase:   phase,
		handler: handler,
	})
}

// Store 返回任务存储
func (r *TaskRunner) Store() *TaskStore {
	return r.store
}

// Start 启动 dispatcher
func (r *TaskRunner) Start(ctx context.Context) {
	ctx, r.cancel = context.WithCancel(ctx)
	r.wg.Add(1)
	go r.dispatcher(ctx)
}

// Stop 优雅关闭
func (r *TaskRunner) Stop() {
	r.cancel()
	r.wg.Wait()
}

// Submit 提交任务
func (r *TaskRunner) Submit(task *model.Task) bool {
	slog.Debug("[taskrunner] 提交任务", "torrent", task.Torrent.Name)
	if !r.store.Add(task) {
		return false // 重复任务
	}
	select {
	case r.queue <- task:
		return true
	default:
		r.store.Remove(task.Torrent.Link)
		return false // 队列满
	}
}

// Cancel 取消任务
func (r *TaskRunner) Cancel(link string) {
	r.store.Remove(link)
}

// dispatcher 从队列取任务，为每个任务启动 goroutine
func (r *TaskRunner) dispatcher(ctx context.Context) {
	slog.Info("[taskrunner] 任务执行器已启动")
	defer r.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-r.queue:
			// 检查任务是否已被取消（从 store 移除）
			if r.store.Get(task.Torrent.Link) == nil {
				continue
			}
			r.wg.Add(1)
			go func() {
				defer r.wg.Done()
				r.process(ctx, task)
			}()
		}
	}
}

// process 处理单个任务的当前阶段
func (r *TaskRunner) process(ctx context.Context, task *model.Task) {
	entry := r.entryFor(task.Phase)
	if entry == nil {
		// 没有 handler 的阶段（如 PhaseCompleted），直接推进
		r.advance(ctx, task)
		return
	}

	// Adding 阶段 acquire 流水线槽位
	if task.Phase == model.PhaseAdding {
		select {
		case r.addingSem <- struct{}{}:
			task.Mu.Lock()
			task.HoldingSlot = true
			task.Mu.Unlock()
		case <-ctx.Done():
			return
		}
	}

	result := entry.handler(ctx, task)

	if result.Err != nil {
		task.Mu.Lock()
		task.Phase = model.PhaseFailed
		task.ErrorMsg = result.Err.Error()
		task.Mu.Unlock()
		r.releaseSlot(task)
		r.store.Remove(task.Torrent.Link)
		slog.Error("[taskrunner] 任务失败",
			"torrent", task.Torrent.Name,
			"phase", task.Phase,
			"error", result.Err)
		return
	}

	if result.PollAfter > 0 {
		// 延迟重入队列，goroutine 立即结束
		time.AfterFunc(result.PollAfter, func() {
			select {
			case r.queue <- task:
			case <-ctx.Done():
			}
		})
		return
	}

	// 成功，推进到下一阶段
	r.advance(ctx, task)
}

// advance 推进到下一阶段
func (r *TaskRunner) advance(ctx context.Context, task *model.Task) {
	task.Mu.Lock()
	oldPhase := task.Phase
	nextPhase := r.nextPhase(task.Phase)
	task.Phase = nextPhase
	task.Mu.Unlock()

	if nextPhase == model.PhaseEnd {
		r.releaseSlot(task)
		r.store.Remove(task.Torrent.Link)
		slog.Info("[taskrunner] 任务完成", "torrent", task.Torrent.Name)
		return
	}

	slog.Debug("[taskrunner] 阶段变更",
		"torrent", task.Torrent.Name,
		"from", oldPhase,
		"to", nextPhase)

	// 立即入队处理下一阶段
	select {
	case r.queue <- task:
	case <-ctx.Done():
	}
}

// releaseSlot 释放流水线槽位
func (r *TaskRunner) releaseSlot(task *model.Task) {
	task.Mu.Lock()
	defer task.Mu.Unlock()
	if task.HoldingSlot {
		<-r.addingSem
		task.HoldingSlot = false
	}
}

// entryFor 查找阶段对应的配置
func (r *TaskRunner) entryFor(phase model.TaskPhase) *phaseEntry {
	for i := range r.phases {
		if r.phases[i].phase == phase {
			return &r.phases[i]
		}
	}
	return nil
}

// nextPhase 返回下一个阶段
func (r *TaskRunner) nextPhase(current model.TaskPhase) model.TaskPhase {
	if current == model.PhaseCompleted {
		return model.PhaseEnd
	}
	return current + 1
}
