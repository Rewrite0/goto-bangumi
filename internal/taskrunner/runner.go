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
	phases []phaseEntry

	// 状态（mu 保护）
	mu            sync.Mutex
	tasks         map[string]*model.Task // 去重 + 查找
	downloadQueue []*model.Task          // Adding/Checking/Downloading 阶段
	generalQueue  []*model.Task          // Renaming 等阶段
	running       int                    // 当前正在执行 handler 的任务数
	downloadSlots int                    // 当前持有下载槽位的任务数

	// 配置
	maxConcurrency int           // 总并发上限
	maxDownload    int           // 下载槽位上限
	slotTimeout    time.Duration // 下载槽位最大持有时间

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
		tasks:          make(map[string]*model.Task),
		maxConcurrency: maxConcurrency,
		maxDownload:    maxDownload,
		slotTimeout:    10 * time.Minute,
		signal:         make(chan struct{}, 1),
	}
}

// Register 注册阶段处理器
func (r *TaskRunner) Register(phase model.TaskPhase, handler PhaseFunc) {
	r.phases = append(r.phases, phaseEntry{
		phase:   phase,
		handler: handler,
	})
}

// needsDownloadSlot 判断阶段是否需要下载槽位
func needsDownloadSlot(phase model.TaskPhase) bool {
	return phase <= model.PhaseDownloading
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
	r.mu.Lock()
	defer r.mu.Unlock()

	if needsDownloadSlot(task.Phase) && (task.HoldingSlot || task.StartTime.IsZero()) {
		r.downloadQueue = append(r.downloadQueue, task)
	} else {
		r.generalQueue = append(r.generalQueue, task)
	}
	r.notify()
}

// finish 任务执行完毕，释放 running 计数
func (r *TaskRunner) finish(task *model.Task) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.running--
	r.notify()
}

// releaseSlot 释放下载槽位
func (r *TaskRunner) releaseSlot(task *model.Task) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if task.HoldingSlot {
		r.downloadSlots--
		task.HoldingSlot = false
	}
	r.notify()
}

// dequeue 从 slice 队列中移除指定索引的任务
func (r *TaskRunner) dequeue(queue *[]*model.Task, idx int) *model.Task {
	q := *queue
	task := q[idx]
	*queue = append(q[:idx], q[idx+1:]...)
	return task
}

// findRunnable 在 downloadQueue 中查找可调度的任务
// 优先找已持有槽位的（HoldingSlot == true），再找需要新槽位的
// 返回索引，-1 表示没有可调度的
func (r *TaskRunner) findRunnable() int {
	// 优先：已持有槽位的任务（sleep 回来的）
	for i, task := range r.downloadQueue {
		if task.HoldingSlot {
			return i
		}
	}
	// 其次：需要新槽位，但槽位未满
	if r.downloadSlots < r.maxDownload {
		if len(r.downloadQueue) > 0 {
			return 0
		}
	}
	return -1
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

// Submit 提交任务
func (r *TaskRunner) Submit(task *model.Task) bool {
	r.mu.Lock()
	link := task.Torrent.Link
	if _, exists := r.tasks[link]; exists {
		r.mu.Unlock()
		slog.Debug("[taskrunner] 任务已存在，忽略", "link", link)
		return false
	}
	r.tasks[link] = task
	r.mu.Unlock()

	slog.Debug("[taskrunner] 提交任务", "torrent", task.Torrent.Name)
	r.enqueue(task)
	return true
}

// Cancel 取消任务
func (r *TaskRunner) Cancel(link string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tasks, link)
}

// Get 根据 link 获取任务
func (r *TaskRunner) Get(link string) *model.Task {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.tasks[link]
}

// remove 从 tasks map 中移除任务
func (r *TaskRunner) remove(link string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tasks, link)
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
	r.mu.Lock()
	defer r.mu.Unlock()

	for {
		scheduled := false

		// 尝试从 downloadQueue 调度
		if r.running < r.maxConcurrency && len(r.downloadQueue) > 0 {
			if idx := r.findRunnable(); idx >= 0 {
				task := r.dequeue(&r.downloadQueue, idx)
				r.dispatch(ctx, task)
				scheduled = true
			}
		}

		// 尝试从 generalQueue 调度
		if r.running < r.maxConcurrency && len(r.generalQueue) > 0 {
			task := r.dequeue(&r.generalQueue, 0)
			r.dispatch(ctx, task)
			scheduled = true
		}

		if !scheduled {
			break
		}
	}
}

// dispatch 启动 goroutine 执行任务（调用方必须持有 mu）
func (r *TaskRunner) dispatch(ctx context.Context, task *model.Task) {
	r.running++
	if task.StartTime.IsZero() {
		task.StartTime = time.Now()
	}
	if !task.HoldingSlot && needsDownloadSlot(task.Phase) {
		r.downloadSlots++
		task.HoldingSlot = true
	}
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		r.process(ctx, task)
	}()
}

// process 处理单个任务
func (r *TaskRunner) process(ctx context.Context, task *model.Task) {
	defer r.finish(task)

	// 检查任务是否已被取消
	if r.Get(task.Torrent.Link) == nil {
		if task.HoldingSlot {
			r.releaseSlot(task)
		}
		return
	}

	// 检查下载槽位是否超时
	if task.HoldingSlot && time.Since(task.StartTime) > r.slotTimeout {
		slog.Info("[taskrunner] 下载槽位超时，释放槽位",
			"torrent", task.Torrent.Name,
			"held", time.Since(task.StartTime))
		r.releaseSlot(task)
	}

	entry := r.entryFor(task.Phase)
	if entry == nil {
		// 没有 handler 的阶段（如 PhaseCompleted），直接推进
		r.advance(ctx, task)
		return
	}

	result := entry.handler(ctx, task)

	if result.Err != nil {
		task.Mu.Lock()
		task.Phase = model.PhaseFailed
		task.ErrorMsg = result.Err.Error()
		task.Mu.Unlock()
		r.releaseSlot(task)
		r.remove(task.Torrent.Link)
		slog.Error("[taskrunner] 任务失败",
			"torrent", task.Torrent.Name,
			"phase", task.Phase,
			"error", result.Err)
		return
	}

	if result.PollAfter > 0 {
		task.Mu.Lock()
		task.RetryCount++
		task.Mu.Unlock()
		// 延迟重入队列，goroutine 立即结束（finish 在 defer 中）
		time.AfterFunc(result.PollAfter, func() {
			if ctx.Err() == nil {
				r.enqueue(task)
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

	// 跨越 Downloading→Renaming 边界时释放下载槽位
	if needsDownloadSlot(oldPhase) && !needsDownloadSlot(nextPhase) {
		r.releaseSlot(task)
	}

	r.enqueue(task)
}
