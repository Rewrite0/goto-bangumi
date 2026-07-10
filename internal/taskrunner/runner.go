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

Task 并发模型
1. scheduler 只有一个 goroutine，但 dispatch 出的不同任务可以在 maxConcurrency 限制下并行执行。
2. 同一个 Task 同一时间最多执行一个 handler。scheduler 选中任务时把状态改为 Queued，worker 开始时
   改为 Running；阶段可以继续时恢复为 Ready，需要延迟轮询时进入 Waiting。
3. Submit、Cancel、handler 完成和轮询定时唤醒可能来自不同 goroutine。tasks 和 downloadSlots
   必须始终在 r.mu 下访问。
4. Task 自身的调度字段由 task.Mutex 保护；同时需要 r.mu 和 task.Mutex 时，固定先获取 r.mu，
   再获取 task.Mutex，避免锁顺序反转。
5. handler 运行期间独占当前 Task，可以直接更新业务字段；其他 goroutine 若要访问同一个 Task，
   必须使用 task.Mutex。Task 引用的 Torrent、Bangumi 不会因为锁住 Task 自动获得并发保护。
6. Cancel 可以和 handler 同时发生：CancelFunc 可以并发调用，但 handler 可能在收到取消后继续收尾。
   当前 tasks 和 downloadSlots 都以 link 为 key；旧 handler 退出前若提交了相同 link 的新 Task，
   因此清理时必须校验 Task 对象身份，不能只按 link 删除。
7. PollAfter 使任务进入 Waiting；定时回调校验 Task 对象身份后把任务改为 Ready，再唤醒 scheduler。
8. 任务状态转换说明：
Created: 初次提交，未被调度
Ready: 当前可以执行，等待 scheduler 调度
Waiting: 等待 PollAfter 到期
Queued: 已被调度，等待 worker 执行
Running: worker 正在执行 handler
Completed: 已结束，不再调度
*/

// TaskRunner 任务执行器
type TaskRunner struct {
	handlers map[model.TaskPhase]PhaseFunc

	// 保护 tasks 和 downloadSlots。
	// 同时需要锁 Task 时，固定先获取 mu，再获取 task.Mutex。
	mu    sync.Mutex
	tasks map[string]*model.Task

	// channel 信号量：len = 当前运行数，cap = 上限
	runningSem chan struct{}

	// 下载槽位由 scheduler 集中分配，key 为 task.Torrent.Link，value 用于区分同 link 的新旧任务。
	maxDownload   int
	downloadSlots map[string]*model.Task
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
		downloadSlots: make(map[string]*model.Task),
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

func (r *TaskRunner) releaseSlotLocked(task *model.Task) {
	link := task.Torrent.Link
	if r.downloadSlots[link] != task {
		return
	}
	delete(r.downloadSlots, link)
}

func (r *TaskRunner) holdingSlotLocked(task *model.Task) bool {
	return r.downloadSlots[task.Torrent.Link] == task
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
	task, ok := r.tasks[link]
	if !ok {
		r.mu.Unlock()
		slog.Debug("[taskrunner] 取消任务失败，任务不存在", "link", link)
		return
	}
	task.Lock()
	task.State = model.TaskStateCompleted
	cancel := task.CancelFunc
	task.Unlock()
	r.removeTaskLocked(task)
	r.mu.Unlock()

	cancel()
	r.notify()
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
// 调度的原则是优先处理有下载槽位且处于可运行状态的任务
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
		task := r.pickTask()
		if task == nil {
			// 到这说明没有可调度的任务了，释放之前占用的并发槽位
			<-r.runningSem
			return
		}
		r.dispatch(task)
	}
}

func (r *TaskRunner) pickTask() *model.Task {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 先取一个正在下载槽位的任务，优先处理它们。
	task := r.pickTaskLocked(func(task *model.Task) bool {
		return r.holdingSlotLocked(task)
	})
	if task != nil {
		return task
	}

	// 取一个下载槽位给还没进入下载流水线的任务。
	if len(r.downloadSlots) < r.maxDownload {
		task = r.pickTaskLocked(func(task *model.Task) bool {
			return needsDownloadSlot(task.CurrentPhase)
		})
		if task != nil {
			r.downloadSlots[task.Torrent.Link] = task
			return task
		}
		// 没有可运行的新下载任务，继续处理一般任务。
	}

	return r.pickTaskLocked(func(task *model.Task) bool {
		return !needsDownloadSlot(task.CurrentPhase)
	})
}

func (r *TaskRunner) pickTaskLocked(match func(*model.Task) bool) *model.Task {
	for _, task := range r.tasks {
		task.Lock()
		if (task.State != model.TaskStateCreated && task.State != model.TaskStateReady) || !match(task) {
			task.Unlock()
			continue
		}
		task.State = model.TaskStateQueued
		task.Unlock()
		return task
	}
	return nil
}

// dispatch 启动 goroutine 执行任务（调用方已持有一个 runningSem 槽位）
func (r *TaskRunner) dispatch(task *model.Task) {
	r.wg.Go(func() {
		defer func() {
			<-r.runningSem
			task.Lock()
			if task.State == model.TaskStateRunning {
				task.State = model.TaskStateReady
			}
			task.Unlock()
			r.notify()
		}()

		task.Lock()
		if task.State != model.TaskStateQueued {
			task.Unlock()
			return
		}
		now := time.Now()
		// 第一次执行任务的时候记录开始时间
		if task.StartTime.IsZero() {
			task.StartTime = now
		}
		if task.EndTime.IsZero() {
			task.EndTime = now.Add(r.slotTimeout)
		}
		ctx := task.Ctx
		task.State = model.TaskStateRunning
		task.Unlock()

		r.process(ctx, task)
	})
}

// process 处理单个任务
// 任务当前阶段没有要处理的 handler 就直接推进到下一阶段
func (r *TaskRunner) process(ctx context.Context, task *model.Task) {
	r.mu.Lock()
	task.Lock()
	slotExpired := r.holdingSlotLocked(task) && time.Now().After(task.EndTime)
	var name string
	var held time.Duration
	if slotExpired {
		name = task.Torrent.Name
		held = time.Since(task.StartTime)
		r.releaseSlotLocked(task)
	}
	task.Unlock()
	r.mu.Unlock()
	if slotExpired {
		slog.Info("[taskrunner] 下载槽位超时，释放槽位",
			"torrent", name,
			"held", held)
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
		task.State = model.TaskStateCompleted
		task.ErrorMsg = result.Err.Error()
		task.Unlock()
		r.removeTaskLocked(task)
		r.mu.Unlock()
		return
	}

	if result.PollAfter > 0 {
		r.mu.Lock()
		task.Lock()
		if r.tasks[task.Torrent.Link] != task || task.State != model.TaskStateRunning {
			task.Unlock()
			r.mu.Unlock()
			return
		}
		task.NextPoll = time.Now().Add(result.PollAfter)
		task.State = model.TaskStateWaiting
		task.Unlock()
		r.mu.Unlock()
		// 到期后将仍在等待的同一个 Task 转为 Ready，再唤醒 scheduler。
		time.AfterFunc(result.PollAfter, func() {
			if r.makeTaskReady(task) {
				r.notify()
			}
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
		task.State = model.TaskStateCompleted
		r.removeTaskLocked(task)
		task.Unlock()
		r.mu.Unlock()
		slog.Debug("[taskrunner] 任务完成", "torrent", task.Torrent.Name)
		return
	}

	if !needsDownloadSlot(nextPhase) {
		r.releaseSlotLocked(task)
	}
	task.Unlock()
	r.mu.Unlock()

	slog.Debug("[taskrunner] 阶段变更",
		"torrent", task.Torrent.Name,
		"from", oldPhase,
		"to", nextPhase)
}

// makeTaskReady 将到期的等待任务转为 Ready。
func (r *TaskRunner) makeTaskReady(task *model.Task) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	link := task.Torrent.Link
	if r.tasks[link] != task {
		return false
	}
	task.Lock()
	defer task.Unlock()
	if task.State != model.TaskStateWaiting {
		return false
	}
	task.State = model.TaskStateReady
	task.NextPoll = time.Time{}
	return true
}

func (r *TaskRunner) removeTaskLocked(task *model.Task) {
	link := task.Torrent.Link
	r.releaseSlotLocked(task)
	if r.tasks[link] == task {
		delete(r.tasks, link)
	}
}

// needsDownloadSlot 判断阶段是否需要下载槽位
func needsDownloadSlot(phase model.TaskPhase) bool {
	return phase <= model.PhaseDownloading
}
