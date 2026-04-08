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

// helper: 创建带 link 的测试 task（从 PhaseAdding 开始）
func newTestTask(link string) *model.Task {
	return &model.Task{
		Phase: model.PhaseAdding,
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

func TestHappyPath_AllPhasesComplete(t *testing.T) {
	addHandler, addCount := successHandler()
	checkHandler, checkCount := successHandler()
	dlHandler, dlCount := successHandler()
	renameHandler, renameCount := successHandler()

	runner := New(16, 2)
	runner.Register(model.PhaseAdding, addHandler)
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

	if v := addCount.Load(); v != 1 {
		t.Errorf("add handler called %d times, want 1", v)
	}
	if v := checkCount.Load(); v != 1 {
		t.Errorf("check handler called %d times, want 1", v)
	}
	if v := dlCount.Load(); v != 1 {
		t.Errorf("download handler called %d times, want 1", v)
	}
	if v := renameCount.Load(); v != 1 {
		t.Errorf("rename handler called %d times, want 1", v)
	}

	if task.Phase != model.PhaseEnd {
		t.Errorf("task phase = %v, want PhaseEnd", task.Phase)
	}

	// 槽位应已释放
	if task.HoldingSlot {
		t.Error("task should not be holding slot after completion")
	}
}

func TestHandlerError_TaskFails(t *testing.T) {
	errBoom := errors.New("boom")

	addHandler, _ := successHandler()
	failHandler := func(ctx context.Context, task *model.Task) PhaseResult {
		return PhaseResult{Err: errBoom}
	}
	dlHandler, dlCount := successHandler()

	runner := New(16, 2)
	runner.Register(model.PhaseAdding, addHandler)
	runner.Register(model.PhaseChecking, failHandler)
	runner.Register(model.PhaseDownloading, dlHandler)

	ctx := context.Background()
	runner.Start(ctx)
	defer runner.Stop()

	task := newTestTask("magnet:fail")
	runner.Submit(task)

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

	if task.Phase != model.PhaseFailed {
		t.Errorf("task phase = %v, want PhaseFailed", task.Phase)
	}
	if task.ErrorMsg != "boom" {
		t.Errorf("task ErrorMsg = %q, want %q", task.ErrorMsg, "boom")
	}
	if v := dlCount.Load(); v != 0 {
		t.Errorf("download handler called %d times, want 0", v)
	}
	// 槽位应已释放（即使任务失败）
	if task.HoldingSlot {
		t.Error("task should not be holding slot after failure")
	}
}

func TestPollAfter_ReEnqueuesTask(t *testing.T) {
	var pollCount atomic.Int32

	addHandler, _ := successHandler()
	pollHandler := func(ctx context.Context, task *model.Task) PhaseResult {
		n := pollCount.Add(1)
		if n < 3 {
			return PhaseResult{PollAfter: 10 * time.Millisecond}
		}
		return PhaseResult{}
	}
	renameHandler, _ := successHandler()

	runner := New(16, 2)
	runner.Register(model.PhaseAdding, addHandler)
	runner.Register(model.PhaseDownloading, pollHandler)
	runner.Register(model.PhaseRenaming, renameHandler)

	ctx := context.Background()
	runner.Start(ctx)
	defer runner.Stop()

	task := newTestTask("magnet:poll")

	runner.Submit(task)

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

	if v := pollCount.Load(); v != 3 {
		t.Errorf("poll handler called %d times, want 3", v)
	}
	if task.Phase != model.PhaseEnd {
		t.Errorf("task phase = %v, want PhaseEnd", task.Phase)
	}
}

func TestDuplicateSubmit_Rejected(t *testing.T) {
	blockCh := make(chan struct{})
	blockHandler := func(ctx context.Context, task *model.Task) PhaseResult {
		<-blockCh
		return PhaseResult{}
	}

	runner := New(16, 2)
	runner.Register(model.PhaseAdding, blockHandler)

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

	task2 := newTestTask("magnet:dup")
	ok2 := runner.Submit(task2)
	if ok2 {
		t.Error("duplicate Submit should return false")
	}
}

func TestCancel_RemovesFromStore(t *testing.T) {
	blockCh := make(chan struct{})
	var called atomic.Int32
	blockHandler := func(ctx context.Context, task *model.Task) PhaseResult {
		called.Add(1)
		<-blockCh
		return PhaseResult{}
	}

	runner := New(16, 2)
	runner.Register(model.PhaseAdding, blockHandler)

	ctx := context.Background()
	runner.Start(ctx)
	defer func() {
		close(blockCh)
		runner.Stop()
	}()

	task := newTestTask("magnet:cancel")
	runner.Submit(task)

	deadline := time.After(3 * time.Second)
	for called.Load() == 0 {
		select {
		case <-deadline:
			t.Fatal("handler was never called")
		case <-time.After(1 * time.Millisecond):
		}
	}

	runner.Cancel("magnet:cancel")

	if runner.Store().Get("magnet:cancel") != nil {
		t.Error("task should be removed from store after Cancel")
	}
}

func TestPipelineSlot_LimitsConcurrency(t *testing.T) {
	const maxConcurrency = 2
	const totalTasks = 5

	var concurrent atomic.Int32
	var maxSeen atomic.Int32
	var wg sync.WaitGroup
	wg.Add(totalTasks)

	// Adding handler 立即成功
	addHandler, _ := successHandler()

	// Downloading handler 模拟耗时操作，用于测量并发
	slowHandler := func(ctx context.Context, task *model.Task) PhaseResult {
		cur := concurrent.Add(1)
		for {
			old := maxSeen.Load()
			if cur <= old || maxSeen.CompareAndSwap(old, cur) {
				break
			}
		}
		time.Sleep(50 * time.Millisecond)
		concurrent.Add(-1)
		wg.Done()
		return PhaseResult{}
	}

	runner := New(64, maxConcurrency)
	runner.Register(model.PhaseAdding, addHandler)
	runner.Register(model.PhaseDownloading, slowHandler)

	ctx := context.Background()
	runner.Start(ctx)
	defer runner.Stop()

	for i := 0; i < totalTasks; i++ {
		task := newTestTask("magnet:sem-" + string(rune('a'+i)))
		runner.Submit(task)
	}

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

	// 因为槽位从 Adding 持有到 PhaseEnd，同时在流水线中的任务不超过 maxConcurrency
	if v := maxSeen.Load(); v > int32(maxConcurrency) {
		t.Errorf("max concurrent = %d, want <= %d", v, maxConcurrency)
	}
	if v := maxSeen.Load(); v < int32(maxConcurrency) {
		t.Logf("warning: max concurrent = %d, expected %d (may be flaky on slow CI)", v, maxConcurrency)
	}
}

func TestReleaseSlot_OnFailure(t *testing.T) {
	// 验证任务失败时槽位被正确释放，不会死锁后续任务
	addHandler, _ := successHandler()

	var failCount atomic.Int32
	failThenSucceed := func(ctx context.Context, task *model.Task) PhaseResult {
		n := failCount.Add(1)
		if n <= 2 {
			return PhaseResult{Err: errors.New("fail")}
		}
		return PhaseResult{}
	}

	runner := New(16, 1) // 只有1个槽位
	runner.Register(model.PhaseAdding, addHandler)
	runner.Register(model.PhaseChecking, failThenSucceed)

	ctx := context.Background()
	runner.Start(ctx)
	defer runner.Stop()

	// 提交3个任务，前2个会在 checking 阶段失败，第3个成功
	for i := 0; i < 3; i++ {
		task := newTestTask("magnet:release-" + string(rune('a'+i)))
		runner.Submit(task)
	}

	// 如果槽位没有正确释放，只有1个槽位会导致后续任务永远拿不到槽位
	deadline := time.After(5 * time.Second)
	for failCount.Load() < 3 {
		select {
		case <-deadline:
			t.Fatalf("only %d tasks processed, expected 3 — slot may not be released on failure", failCount.Load())
		case <-time.After(10 * time.Millisecond):
		}
	}
}
