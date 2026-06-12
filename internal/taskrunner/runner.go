package taskrunner

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"goto-bangumi/internal/model"
)

/*
整体的一个思路说明
一个任务有几个阶段, 每个阶段有对应的 handler, 目前每个阶段只有一个 handler
所有和下载器有关的交互都要提交任务来运行，不允许自己跑
下载任务只能有 n 个，下载槽位有超时机制，超过时间会被强制释放，让其他的任务有机会下载
调度优先看有下载槽位的任务，如果都在休息，就看一般任务
每个任务可以被单独取消，删除所有的调度
任务超时失败的任务也不再处理
*/

// TaskRunner 任务执行器
type TaskRunner struct {
	handlers map[model.TaskPhase]PhaseFunc

	// tasks map 的保护锁
	mu    sync.Mutex
	tasks map[string]*model.Task

	queue taskQueue

	// channel 信号量：len = 当前占用，cap = 上限
	runningSem  chan struct{}
	downloadSem chan struct{}
	slotTimeout time.Duration

	// 控制
	signal chan struct{} // buffer 1，唤醒 scheduler
	wg     sync.WaitGroup
	cancel context.CancelFunc
}

// New 创建任务执行器
func New(maxConcurrency, maxDownload int) *TaskRunner {
	if maxConcurrency <= 0 {
		maxConcurrency = 4
	}
	if maxDownload <= 0 {
		maxDownload = 5
	}
	return &TaskRunner{
		handlers:    make(map[model.TaskPhase]PhaseFunc),
		tasks:       make(map[string]*model.Task),
		runningSem:  make(chan struct{}, maxConcurrency),
		downloadSem: make(chan struct{}, maxDownload),
		slotTimeout: 10 * time.Minute,
		signal:      make(chan struct{}, 1),
	}
}

// Register 注册阶段处理器
func (r *TaskRunner) Register(phase model.TaskPhase, handler PhaseFunc) {
	if r.handlers == nil {
		r.handlers = make(map[model.TaskPhase]PhaseFunc)
	}
	r.handlers[phase] = handler
}

// notify 非阻塞写入 signal
func (r *TaskRunner) notify() {
	select {
	case r.signal <- struct{}{}:
	default:
	}
}

// enqueue 根据阶段放入对应队列
func (r *TaskRunner) enqueue(task *model.Task) {
	r.queue.enqueue(task)
	r.notify()
}

// releaseSlot 释放下载槽位
func (r *TaskRunner) releaseSlot(task *model.Task) {
	if task.HoldingSlot {
		<-r.downloadSem
		task.HoldingSlot = false
		r.notify()
	}
}

// remove 从 tasks 中删除任务
func (r *TaskRunner) remove(link string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tasks, link)
}

// handlerFor 查找阶段对应的处理器
func (r *TaskRunner) handlerFor(phase model.TaskPhase) PhaseFunc {
	return r.handlers[phase]
}

// nextPhase 返回下一个阶段
func (r *TaskRunner) nextPhase(current model.TaskPhase) model.TaskPhase {
	if current >= model.PhaseCompleted {
		return model.PhaseEnd
	}
	return current + 1
}

// Submit 提交任务
func (r *TaskRunner) Submit(task *model.Task) bool {
	r.mu.Lock()
	link := task.Torrent.Link
	if _, exists := r.tasks[link]; exists {
		slog.Debug("[taskrunner] 任务已存在，忽略", "link", link)
		r.mu.Unlock()
		return false
	}
	r.tasks[link] = task
	slog.Debug("[taskrunner] 提交任务", "torrent", task.Torrent.Name)
	r.mu.Unlock()
	r.enqueue(task)
	return true
}

// Cancel 取消任务
func (r *TaskRunner) Cancel(link string) {
	r.mu.Lock()
	task := r.tasks[link]
	delete(r.tasks, link)
	r.mu.Unlock()
	if task != nil {
		r.queue.remove(task)
	}
}

// Get 根据 link 获取任务
func (r *TaskRunner) Get(link string) *model.Task {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.tasks[link]
}

// Start 启动 scheduler
func (r *TaskRunner) Start(ctx context.Context) {
	ctx, r.cancel = context.WithCancel(ctx)
	r.wg.Add(1)
	go r.scheduler(ctx)
}

// Stop 优雅关闭
func (r *TaskRunner) Stop() {
	r.cancel()
	r.wg.Wait()
}

// scheduler 事件驱动调度循环
func (r *TaskRunner) scheduler(ctx context.Context) {
	slog.Info("[taskrunner] 任务执行器已启动")
	defer r.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case <-r.signal:
			r.schedule(ctx)
		}
	}
}

// schedule 尝试尽可能多地调度任务
func (r *TaskRunner) schedule(ctx context.Context) {
	for {
		// 非阻塞获取并发槽位
		select {
		case r.runningSem <- struct{}{}:
		default:
			return
		}

		task := r.queue.tryDequeueDownload(len(r.downloadSem), cap(r.downloadSem))
		if task == nil {
			task = r.queue.tryDequeueGeneral()
		}
		if task == nil {
			<-r.runningSem
			return
		}

		r.dispatch(ctx, task)
	}
}

// dispatch 启动 goroutine 执行任务（调用方已持有一个 runningSem 槽位）
func (r *TaskRunner) dispatch(ctx context.Context, task *model.Task) {
	if task.StartTime.IsZero() {
		task.StartTime = time.Now()
	}
	if task.EndTime.IsZero() {
		task.EndTime = time.Now().Add(r.slotTimeout)
	}
	if !task.HoldingSlot && needsDownloadSlot(task.CurrentPhase) {
		// schedule() 已验证 len(downloadSem) < cap，此处不会阻塞
		r.downloadSem <- struct{}{}
		task.HoldingSlot = true
	}
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		defer func() {
			<-r.runningSem
			r.notify()
		}()
		r.process(ctx, task)
	}()
}

// process 处理单个任务
func (r *TaskRunner) process(ctx context.Context, task *model.Task) {
	if task.HoldingSlot && time.Now().After(task.EndTime) {
		slog.Info("[taskrunner] 下载槽位超时，释放槽位",
			"torrent", task.Torrent.Name,
			"held", time.Since(task.StartTime))
		r.releaseSlot(task)
	}

	handler := r.handlerFor(task.CurrentPhase)
	if handler == nil {
		r.advance(task)
		return
	}

	result := handler(ctx, task)

	if result.Err != nil {
		slog.Error("[taskrunner] 任务失败",
			"torrent", task.Torrent.Name,
			"phase", task.CurrentPhase,
			"error", result.Err)
		task.CurrentPhase = model.PhaseFailed
		task.ErrorMsg = result.Err.Error()
		r.releaseSlot(task)
		r.remove(task.Torrent.Link)
		return
	}

	if result.PollAfter > 0 {
		task.RetryCount++
		// context.AfterFunc 返回的 stop 函数：在 timer 触发时调用 stop()，
		// 若返回 true 说明 ctx 尚未取消，可以安全入队；返回 false 则跳过。
		stop := context.AfterFunc(ctx, func() {})
		time.AfterFunc(result.PollAfter, func() {
			if stop() {
				r.enqueue(task)
			}
		})
		return
	}

	r.advance(task)
}

// advance 推进到下一阶段
func (r *TaskRunner) advance(task *model.Task) {
	task.Mu.Lock()
	oldPhase := task.CurrentPhase
	nextPhase := r.nextPhase(task.CurrentPhase)
	task.CurrentPhase = nextPhase
	task.RetryCount = 0
	task.Mu.Unlock()

	if nextPhase == model.PhaseEnd {
		r.releaseSlot(task)
		r.remove(task.Torrent.Link)
		slog.Info("[taskrunner] 任务完成", "torrent", task.Torrent.Name)
		return
	}

	slog.Debug("[taskrunner] 阶段变更",
		"torrent", task.Torrent.Name,
		"from", oldPhase,
		"to", nextPhase)

	if needsDownloadSlot(oldPhase) && !needsDownloadSlot(nextPhase) {
		r.releaseSlot(task)
	}
	r.enqueue(task)
}

// needsDownloadSlot 判断阶段是否需要下载槽位
func needsDownloadSlot(phase model.TaskPhase) bool {
	return phase <= model.PhaseDownloading
}
