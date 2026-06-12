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

	// queue taskQueue
	// downloadqueue 里面只有满足调度的任务, 到重试的任务到时间才会重新加入队列
	// 所以如果下载槽位有需求,可以直接拿出来调度
	downloadQueue taskQueue
	generalQueue  taskQueue

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
// maxDownload 并不代表最多有几个下载, 只是尽量保证, 是为了防止过慢
// 的任务一直持有, 导致其他任务没有机会下载了
func New(maxConcurrency, maxDownload int) *TaskRunner {
	if maxConcurrency <= 0 {
		maxConcurrency = 10
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
	// 需要下载槽位的任务如果已经持有或者还没开始过，就放下载队列，其他的放一般队列
	if needsDownloadSlot(task.CurrentPhase) && (task.HoldingSlot || task.StartTime.IsZero()) {
		r.downloadQueue.enqueue(task)
	} else {
		r.generalQueue.enqueue(task)
	}
	r.notify()
}

// releaseSlot 释放下载槽位
// 入队后通知调度一次
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

// Submit 提交任务
func (r *TaskRunner) Submit(task *model.Task) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	link := task.Torrent.Link

	if _, exists := r.tasks[link]; exists {
		slog.Debug("[taskrunner] 任务已存在，忽略", "link", link)
		return false
	}
	ctx, cancel := context.WithCancel(context.Background())
	task.CancelFunc = cancel
	task.Ctx = ctx
	r.tasks[link] = task
	slog.Debug("[taskrunner] 提交任务", "torrent", task.Torrent.Name)
	r.enqueue(task)
	return true
}

// Cancel 取消任务
func (r *TaskRunner) Cancel(link string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	task, ok := r.tasks[link]
	if !ok {
		slog.Debug("[taskrunner] 取消任务失败，任务不存在", "link", link)
		return
	}

	task.CancelFunc()
	r.downloadQueue.remove(task)
	r.generalQueue.remove(task)
	delete(r.tasks, link)
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

// Stop 关闭所有任务
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

// schedule 尝试尽可能多地调度任务, 除非没有可调度的任务了,不然会一直调度下去
// 先看看正在下载的任务, 有没有要处理的
// 调度的原则是优先处理有下载槽位的任务(要求满足时间)
// 若还有空闲的下载槽位, 就从下载队列里取一个任务来处理
// 如果上面没有要开始的任务, 就处理一般任务
func (r *TaskRunner) schedule(ctx context.Context) {
	for {
		// 非阻塞获取并发槽位
		select {
		case r.runningSem <- struct{}{}:
		default:
			return
		}
		var task *model.Task
		// 取一个下载槽位
		select {
		case r.downloadSem <- struct{}{}:
			task = r.downloadQueue.tryDequeue()
			if task != nil {
				// 已经执有的返回现在拿到的位置
				if task.HoldingSlot{
					<-r.downloadSem
				}
				task.HoldingSlot = true
			} else {
			// 空的任务说明没有满足条件的任务了，先释放下载槽位
				<-r.downloadSem
				task = r.generalQueue.tryDequeue()
			}
		default:
			// 到这说明下载槽位满了，先处理一般任务
			task = r.generalQueue.tryDequeue()
		}
		if task == nil {
			// 到这说明没有可调度的任务了，释放之前占用的并发槽位
			<-r.runningSem
			return
		}
		r.dispatch(task.Ctx, task)
	}
}

// dispatch 启动 goroutine 执行任务（调用方已持有一个 runningSem 槽位）
func (r *TaskRunner) dispatch(ctx context.Context, task *model.Task) {
	// 第一次执行任务的时候记录开始时间
	if task.StartTime.IsZero() {
		task.StartTime = time.Now()
	}
	if task.EndTime.IsZero() {
		task.EndTime = time.Now().Add(r.slotTimeout)
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
// 任务当前阶段没有要处理的 handler 就直接推进到下一阶段
func (r *TaskRunner) process(ctx context.Context, task *model.Task) {
	if task.HoldingSlot && time.Now().After(task.EndTime) {
		slog.Info("[taskrunner] 下载槽位超时，释放槽位",
			"torrent", task.Torrent.Name,
			"held", time.Since(task.StartTime))
		r.releaseSlot(task)
	}

	handler := r.handlers[task.CurrentPhase]
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
		// 重试的任务在这里入队
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
// 现在是按阶段推进, 后序再考虑阶段跳转的情况
func (r *TaskRunner) advance(task *model.Task) {
	oldPhase := task.CurrentPhase
	nextPhase := oldPhase + 1
	// FAIL->END, COMPLETED->END
	if oldPhase >= model.PhaseCompleted {
		nextPhase = model.PhaseEnd
	}
	task.CurrentPhase = nextPhase
	task.RetryCount = 0

	if nextPhase == model.PhaseEnd {
		r.releaseSlot(task)
		r.remove(task.Torrent.Link)
		slog.Debug("[taskrunner] 任务完成", "torrent", task.Torrent.Name)
		return
	}

	slog.Debug("[taskrunner] 阶段变更",
		"torrent", task.Torrent.Name,
		"from", oldPhase,
		"to", nextPhase)

	if task.HoldingSlot && !needsDownloadSlot(nextPhase) {
		r.releaseSlot(task)
	}
	r.enqueue(task)
}

// needsDownloadSlot 判断阶段是否需要下载槽位
func needsDownloadSlot(phase model.TaskPhase) bool {
	return phase <= model.PhaseDownloading
}
