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

	// channel 信号量：len = 当前运行数，cap = 上限
	runningSem chan struct{}

	// 下载槽位由 scheduler 集中分配，key 为 task.Torrent.Link
	maxDownload   int
	downloadSlots map[string]struct{}
	slotTimeout   time.Duration

	// 控制
	signal chan struct{} // buffer 1，唤醒 scheduler
	wg     sync.WaitGroup
	cancel context.CancelFunc
}

// New 创建任务执行器。
// maxDownload 表示最多同时持有下载流水线槽位的任务数；槽位可能跨 Adding/Checking/Downloading 多个阶段持有。
// 默认的并行任务数是 8, 下载槽位数是 4
func New(maxConcurrency, maxDownload int) *TaskRunner {
	return &TaskRunner{
		handlers:      make(map[model.TaskPhase]PhaseFunc),
		tasks:         make(map[string]*model.Task),
		runningSem:    make(chan struct{}, maxConcurrency),
		maxDownload:   maxDownload,
		downloadSlots: make(map[string]struct{}),
		slotTimeout:   10 * time.Minute,
		signal:        make(chan struct{}, 1),
	}
}

// Register 注册阶段处理器
func (r *TaskRunner) Register(phase model.TaskPhase, handler PhaseFunc) {
	r.handlers[phase] = handler
}

// notify 非阻塞写入 signal
func (r *TaskRunner) notify() {
	select {
	case r.signal <- struct{}{}:
	default:
	}
}

// releaseSlot 释放下载槽位
// 释放后通知调度一次
// 在取消，阶段转换下都有可能调用
func (r *TaskRunner) releaseSlot(task *model.Task) {
	r.mu.Lock()
	released := r.releaseSlotLocked(task)
	r.mu.Unlock()
	if released {
		r.notify()
	}
}

func (r *TaskRunner) releaseSlotLocked(task *model.Task) bool {
	link := task.Torrent.Link
	if _, ok := r.downloadSlots[link]; !ok {
		return false
	}
	delete(r.downloadSlots, link)
	return true
}

func (r *TaskRunner) holdingSlotLocked(task *model.Task) bool {
	_, ok := r.downloadSlots[task.Torrent.Link]
	return ok
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
	task.Lock()
	task.CancelFunc = cancel
	task.Ctx = ctx
	task.Unlock()
	r.tasks[link] = task
	slog.Debug("[taskrunner] 提交任务", "torrent", task.Torrent.Name)
	r.notify()
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
	released := r.removeTaskLocked(task)
	if released {
		r.notify()
	}
}

// Start 启动 scheduler
func (r *TaskRunner) Start(ctx context.Context) {
	ctx, r.cancel = context.WithCancel(ctx)
	r.wg.Go(func() {
		r.scheduler(ctx)
	})
}

// Stop 关闭所有任务
func (r *TaskRunner) Stop() {
	r.cancel()
	r.wg.Wait()
}

// scheduler 事件驱动调度循环
func (r *TaskRunner) scheduler(ctx context.Context) {
	slog.Info("[taskrunner] 任务执行器已启动")
	for {
		select {
		case <-ctx.Done():
			return
		case <-r.signal:
			r.schedule()
		}
	}
}

// schedule 尝试尽可能多地调度任务, 除非没有可调度的任务了,不然会一直调度下去
// 先看看正在下载的任务, 有没有要处理的
// 调度的原则是优先处理有下载槽位的任务(要求满足时间)
// 若还有空闲的下载槽位, 就从 tasks 里找一个下载任务来处理
// 如果上面没有要开始的任务, 就处理一般任务
// TODO: 任务应该有一个时间限制
func (r *TaskRunner) schedule() {
	for {
		// 非阻塞获取并发槽位
		select {
		case r.runningSem <- struct{}{}:
		default:
			return
		}
		task := r.pickTask(time.Now())
		if task == nil {
			// 到这说明没有可调度的任务了，释放之前占用的并发槽位
			<-r.runningSem
			return
		}
		r.dispatch(task)
	}
}

func (r *TaskRunner) pickTask(now time.Time) *model.Task {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 先取一个正在下载槽位的任务，优先处理它们。
	task := r.pickTaskLocked(now, func(task *model.Task) bool {
		return r.holdingSlotLocked(task)
	})
	if task != nil {
		return task
	}

	// 取一个下载槽位给还没进入下载流水线的任务。
	if len(r.downloadSlots) < r.maxDownload {
		task = r.pickTaskLocked(now, func(task *model.Task) bool {
			return needsDownloadSlot(task.CurrentPhase) && !r.holdingSlotLocked(task)
		})
		if task != nil {
			r.downloadSlots[task.Torrent.Link] = struct{}{}
			return task
		}
		// 到这说明下载槽位满了，先处理一般任务。
	}

	return r.pickTaskLocked(now, func(task *model.Task) bool {
		return !needsDownloadSlot(task.CurrentPhase)
	})
}

func (r *TaskRunner) pickTaskLocked(now time.Time, match func(*model.Task) bool) *model.Task {
	for _, task := range r.tasks {
		task.Lock()
		if !isRunnable(task, now) || !match(task) {
			task.Unlock()
			continue
		}
		task.State = model.TaskStateRunning
		task.NextPoll = time.Time{}
		task.Unlock()
		return task
	}
	return nil
}

func isRunnable(task *model.Task, now time.Time) bool {
	if task.State != model.TaskStateCreated && task.State != model.TaskStateIdle {
		return false
	}
	return task.NextPoll.IsZero() || !now.Before(task.NextPoll)
}

// dispatch 启动 goroutine 执行任务（调用方已持有一个 runningSem 槽位）
func (r *TaskRunner) dispatch(task *model.Task) {
	task.Lock()
	now := time.Now()
	// 第一次执行任务的时候记录开始时间
	if task.StartTime.IsZero() {
		task.StartTime = now
	}
	if task.EndTime.IsZero() {
		task.EndTime = now.Add(r.slotTimeout)
	}
	ctx := task.Ctx
	task.Unlock()

	r.wg.Go(func() {
		defer func() {
			<-r.runningSem
			r.notify()
		}()
		r.process(ctx, task)
	})
}

// process 处理单个任务
// 任务当前阶段没有要处理的 handler 就直接推进到下一阶段
func (r *TaskRunner) process(ctx context.Context, task *model.Task) {
	r.mu.Lock()
	task.Lock()
	if r.holdingSlotLocked(task) && time.Now().After(task.EndTime) {
		name := task.Torrent.Name
		held := time.Since(task.StartTime)
		released := r.releaseSlotLocked(task)
		task.Unlock()
		r.mu.Unlock()
		slog.Info("[taskrunner] 下载槽位超时，释放槽位",
			"torrent", name,
			"held", held)
		if released {
			r.notify()
		}
	} else {
		task.Unlock()
		r.mu.Unlock()
	}

	task.Lock()
	phase := task.CurrentPhase
	task.Unlock()
	handler := r.handlers[phase]
	if handler == nil {
		r.advance(task)
		return
	}

	result := handler(ctx, task)

	if result.Err != nil {
		slog.Error("[taskrunner] 任务失败",
			"torrent", task.Torrent.Name,
			"phase", phase,
			"error", result.Err)
		r.mu.Lock()
		task.Lock()
		task.CurrentPhase = model.PhaseFailed
		task.ErrorMsg = result.Err.Error()
		task.Unlock()
		released := r.removeTaskLocked(task)
		r.mu.Unlock()
		if released {
			r.notify()
		}
		return
	}

	if result.PollAfter > 0 {
		task.Lock()
		task.RetryCount++
		task.NextPoll = time.Now().Add(result.PollAfter)
		task.State = model.TaskStateIdle
		task.Unlock()
		// 重试的任务到时间后只唤醒调度器，由调度器扫描 tasks 决定是否可运行。
		time.AfterFunc(result.PollAfter, func() {
			select {
			case <-ctx.Done():
				return
			default:
			}
			r.notify()
		})
		return
	}
	r.advance(task)
}

// advance 推进到下一阶段
// 现在是按阶段推进, 后序再考虑阶段跳转的情况
func (r *TaskRunner) advance(task *model.Task) {
	r.mu.Lock()
	task.Lock()
	oldPhase := task.CurrentPhase
	nextPhase := oldPhase + 1
	// FAIL->END, COMPLETED->END
	if oldPhase >= model.PhaseCompleted {
		nextPhase = model.PhaseEnd
	}
	task.CurrentPhase = nextPhase
	task.RetryCount = 0
	task.NextPoll = time.Time{}

	if nextPhase == model.PhaseEnd {
		released := r.removeTaskLocked(task)
		task.Unlock()
		r.mu.Unlock()
		slog.Debug("[taskrunner] 任务完成", "torrent", task.Torrent.Name)
		if released {
			r.notify()
		}
		return
	}

	released := false
	if !needsDownloadSlot(nextPhase) {
		released = r.releaseSlotLocked(task)
	}
	task.State = model.TaskStateIdle
	task.Unlock()
	r.mu.Unlock()

	slog.Debug("[taskrunner] 阶段变更",
		"torrent", task.Torrent.Name,
		"from", oldPhase,
		"to", nextPhase)
	if released {
		r.notify()
	}
	r.notify()
}

func (r *TaskRunner) removeTaskLocked(task *model.Task) bool {
	released := r.releaseSlotLocked(task)
	delete(r.tasks, task.Torrent.Link)
	return released
}

// needsDownloadSlot 判断阶段是否需要下载槽位
func needsDownloadSlot(phase model.TaskPhase) bool {
	return phase <= model.PhaseDownloading
}
