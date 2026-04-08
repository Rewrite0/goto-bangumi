package taskrunner

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"goto-bangumi/internal/model"
)

// helper: 创建带 link 的测试 task
func newTestTask(link string) *model.Task {
	return &model.Task{
		Phase: model.PhaseChecking,
		Torrent: &model.Torrent{
			Link: link,
			Name: "test-" + link,
		},
		Bangumi:   &model.Bangumi{},
		StartTime: time.Now(),
	}
}

// helper: 创建一个记录调用次数的 handler，成功返回
func successHandler() (PhaseFunc, *atomic.Int32) {
	var count atomic.Int32
	return func(ctx context.Context, task *model.Task) PhaseResult {
		count.Add(1)
		return PhaseResult{}
	}, &count
}

func TestHappyPath_AllPhasesComplete(t *testing.T) {
	// 三个阶段各有一个 handler，全部成功
	checkHandler, checkCount := successHandler()
	dlHandler, dlCount := successHandler()
	renameHandler, renameCount := successHandler()

	runner := New(16, 2)
	runner.Register(model.PhaseChecking, checkHandler)
	runner.Register(model.PhaseDownloading, dlHandler)
	runner.Register(model.PhaseRenaming, renameHandler)

	ctx := context.Background()
	runner.Start(ctx)
	defer runner.Stop()

	task := newTestTask("magnet:happy")
	ok := runner.Submit(task)
	if !ok {
		t.Fatal("Submit should succeed")
	}

	// 等待任务完成：task 从 store 移除表示已完成
	deadline := time.After(3 * time.Second)
	for {
		if runner.Store().Get("magnet:happy") == nil {
			break
		}
		select {
		case <-deadline:
			t.Fatal("task did not complete within deadline")
		case <-time.After(10 * time.Millisecond):
		}
	}

	// 每个 handler 应被调用恰好 1 次
	if v := checkCount.Load(); v != 1 {
		t.Errorf("check handler called %d times, want 1", v)
	}
	if v := dlCount.Load(); v != 1 {
		t.Errorf("download handler called %d times, want 1", v)
	}
	if v := renameCount.Load(); v != 1 {
		t.Errorf("rename handler called %d times, want 1", v)
	}

	// 任务最终阶段应为 PhaseEnd
	if task.Phase != model.PhaseEnd {
		t.Errorf("task phase = %v, want PhaseEnd", task.Phase)
	}
}

func TestHandlerError_TaskFails(t *testing.T) {
	errBoom := errors.New("boom")

	// check handler 返回错误
	failHandler := func(ctx context.Context, task *model.Task) PhaseResult {
		return PhaseResult{Err: errBoom}
	}
	dlHandler, dlCount := successHandler()

	runner := New(16, 2)
	runner.Register(model.PhaseChecking, failHandler)
	runner.Register(model.PhaseDownloading, dlHandler)

	ctx := context.Background()
	runner.Start(ctx)
	defer runner.Stop()

	task := newTestTask("magnet:fail")
	runner.Submit(task)

	// 等待任务从 store 移除
	deadline := time.After(3 * time.Second)
	for {
		if runner.Store().Get("magnet:fail") == nil {
			break
		}
		select {
		case <-deadline:
			t.Fatal("failed task was not removed from store")
		case <-time.After(10 * time.Millisecond):
		}
	}

	// 任务应标记为 PhaseFailed
	if task.Phase != model.PhaseFailed {
		t.Errorf("task phase = %v, want PhaseFailed", task.Phase)
	}

	// ErrorMsg 应包含错误信息
	if task.ErrorMsg != "boom" {
		t.Errorf("task ErrorMsg = %q, want %q", task.ErrorMsg, "boom")
	}

	// 下载 handler 不应被调用（任务在 check 阶段就失败了）
	if v := dlCount.Load(); v != 0 {
		t.Errorf("download handler called %d times, want 0", v)
	}
}

func TestPollAfter_ReEnqueuesTask(t *testing.T) {
	var pollCount atomic.Int32

	// 前两次返回 PollAfter，第三次成功
	pollHandler := func(ctx context.Context, task *model.Task) PhaseResult {
		n := pollCount.Add(1)
		if n < 3 {
			return PhaseResult{PollAfter: 10 * time.Millisecond}
		}
		return PhaseResult{}
	}
	renameHandler, _ := successHandler()

	runner := New(16, 2)
	// 跳过 checking，直接从 downloading 开始
	runner.Register(model.PhaseDownloading, pollHandler)
	runner.Register(model.PhaseRenaming, renameHandler)

	ctx := context.Background()
	runner.Start(ctx)
	defer runner.Stop()

	task := newTestTask("magnet:poll")
	task.Phase = model.PhaseDownloading // 直接从 downloading 阶段开始

	runner.Submit(task)

	// 等待任务完成
	deadline := time.After(3 * time.Second)
	for {
		if runner.Store().Get("magnet:poll") == nil {
			break
		}
		select {
		case <-deadline:
			t.Fatal("task did not complete within deadline")
		case <-time.After(10 * time.Millisecond):
		}
	}

	// pollHandler 应被调用 3 次（2 次 poll + 1 次成功）
	if v := pollCount.Load(); v != 3 {
		t.Errorf("poll handler called %d times, want 3", v)
	}

	if task.Phase != model.PhaseEnd {
		t.Errorf("task phase = %v, want PhaseEnd", task.Phase)
	}
}

func TestDuplicateSubmit_Rejected(t *testing.T) {
	// 用一个会阻塞的 handler，确保第一个任务还在处理中
	blockCh := make(chan struct{})
	blockHandler := func(ctx context.Context, task *model.Task) PhaseResult {
		<-blockCh
		return PhaseResult{}
	}

	runner := New(16, 2)
	runner.Register(model.PhaseChecking, blockHandler)

	ctx := context.Background()
	runner.Start(ctx)
	defer func() {
		close(blockCh)
		runner.Stop()
	}()

	task1 := newTestTask("magnet:dup")
	ok1 := runner.Submit(task1)
	if !ok1 {
		t.Fatal("first Submit should succeed")
	}

	// 再次提交相同 link
	task2 := newTestTask("magnet:dup")
	ok2 := runner.Submit(task2)
	if ok2 {
		t.Error("duplicate Submit should return false")
	}
}

func TestCancel_RemovesFromStore(t *testing.T) {
	// handler 会阻塞直到被通知
	blockCh := make(chan struct{})
	var called atomic.Int32
	blockHandler := func(ctx context.Context, task *model.Task) PhaseResult {
		called.Add(1)
		<-blockCh
		return PhaseResult{}
	}

	runner := New(16, 2)
	runner.Register(model.PhaseChecking, blockHandler)

	ctx := context.Background()
	runner.Start(ctx)
	defer func() {
		close(blockCh)
		runner.Stop()
	}()

	task := newTestTask("magnet:cancel")
	runner.Submit(task)

	// 等 handler 被调用
	deadline := time.After(3 * time.Second)
	for called.Load() == 0 {
		select {
		case <-deadline:
			t.Fatal("handler was never called")
		case <-time.After(1 * time.Millisecond):
		}
	}

	// Cancel
	runner.Cancel("magnet:cancel")

	// 应从 store 移除
	if runner.Store().Get("magnet:cancel") != nil {
		t.Error("task should be removed from store after Cancel")
	}
}

func TestSemaphore_LimitsConcurrency(t *testing.T) {
	const maxConcurrency = 2
	const totalTasks = 5

	var concurrent atomic.Int32
	var maxSeen atomic.Int32
	var wg sync.WaitGroup
	wg.Add(totalTasks)

	slowHandler := func(ctx context.Context, task *model.Task) PhaseResult {
		cur := concurrent.Add(1)
		// 记录最大并发数
		for {
			old := maxSeen.Load()
			if cur <= old || maxSeen.CompareAndSwap(old, cur) {
				break
			}
		}
		time.Sleep(50 * time.Millisecond) // 模拟耗时操作
		concurrent.Add(-1)
		wg.Done()
		return PhaseResult{}
	}

	runner := New(64, maxConcurrency)
	runner.Register(model.PhaseChecking, slowHandler)

	ctx := context.Background()
	runner.Start(ctx)
	defer runner.Stop()

	for i := 0; i < totalTasks; i++ {
		task := newTestTask("magnet:sem-" + string(rune('a'+i)))
		runner.Submit(task)
	}

	// 等待所有 handler 完成
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("tasks did not complete within deadline")
	}

	if v := maxSeen.Load(); v > int32(maxConcurrency) {
		t.Errorf("max concurrent = %d, want <= %d", v, maxConcurrency)
	}
	// 至少应达到 maxConcurrency（5个任务，2个并发，应该能同时跑2个）
	if v := maxSeen.Load(); v < int32(maxConcurrency) {
		t.Logf("warning: max concurrent = %d, expected %d (may be flaky on slow CI)", v, maxConcurrency)
	}
}

func TestNew_DirectParams(t *testing.T) {
	runner := New(16, 2)
	if runner == nil {
		t.Fatal("New should return non-nil runner")
	}
	if cap(runner.queue) != 16 {
		t.Errorf("queue cap = %d, want 16", cap(runner.queue))
	}
	if cap(runner.addingSem) != 2 {
		t.Errorf("addingSem cap = %d, want 2", cap(runner.addingSem))
	}
}
