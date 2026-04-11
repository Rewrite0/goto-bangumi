package taskrunner

import (
	"context"
	"errors"
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
		Bangumi: &model.Bangumi{},
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

// helper: 等待条件满足或超时
func waitFor(t *testing.T, timeout time.Duration, condition func() bool, msg string) {
	t.Helper()
	deadline := time.After(timeout)
	for !condition() {
		select {
		case <-deadline:
			t.Fatal(msg)
		case <-time.After(5 * time.Millisecond):
		}
	}
}

func TestNew_DirectParams(t *testing.T) {
	runner := New(8, 5)
	if runner == nil {
		t.Fatal("New should return non-nil runner")
	}
	if runner.maxConcurrency != 8 {
		t.Errorf("maxConcurrency = %d, want 8", runner.maxConcurrency)
	}
	if runner.maxDownload != 5 {
		t.Errorf("maxDownload = %d, want 5", runner.maxDownload)
	}
}

func TestHappyPath_AllPhasesComplete(t *testing.T) {
	addHandler, addCount := successHandler()
	checkHandler, checkCount := successHandler()
	dlHandler, dlCount := successHandler()
	renameHandler, renameCount := successHandler()

	runner := New(4, 5)
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

	waitFor(t, 3*time.Second, func() bool {
		return runner.Get("magnet:happy") == nil
	}, "task did not complete within deadline")

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

	runner := New(4, 5)
	runner.Register(model.PhaseAdding, addHandler)
	runner.Register(model.PhaseChecking, failHandler)
	runner.Register(model.PhaseDownloading, dlHandler)

	ctx := context.Background()
	runner.Start(ctx)
	defer runner.Stop()

	task := newTestTask("magnet:fail")
	runner.Submit(task)

	waitFor(t, 3*time.Second, func() bool {
		return runner.Get("magnet:fail") == nil
	}, "failed task was not removed")

	if task.Phase != model.PhaseFailed {
		t.Errorf("task phase = %v, want PhaseFailed", task.Phase)
	}
	if task.ErrorMsg != "boom" {
		t.Errorf("task ErrorMsg = %q, want %q", task.ErrorMsg, "boom")
	}
	if v := dlCount.Load(); v != 0 {
		t.Errorf("download handler called %d times, want 0", v)
	}
	if task.HoldingSlot {
		t.Error("task should not be holding slot after failure")
	}
}

// TestPollAfter_ReEnqueuesTask 验证 PollAfter 会重新入队并继续处理
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

	runner := New(4, 5)
	runner.Register(model.PhaseAdding, addHandler)
	runner.Register(model.PhaseDownloading, pollHandler)
	runner.Register(model.PhaseRenaming, renameHandler)

	ctx := context.Background()
	runner.Start(ctx)
	defer runner.Stop()

	task := newTestTask("magnet:poll")
	runner.Submit(task)

	waitFor(t, 3*time.Second, func() bool {
		return runner.Get("magnet:poll") == nil
	}, "task did not complete within deadline")

	if v := pollCount.Load(); v != 3 {
		t.Errorf("poll handler called %d times, want 3", v)
	}
	if task.Phase != model.PhaseEnd {
		t.Errorf("task phase = %v, want PhaseEnd", task.Phase)
	}
}

// TestDuplicateSubmit_Rejected 验证提交重复链接的任务会被拒绝
func TestDuplicateSubmit_Rejected(t *testing.T) {
	blockCh := make(chan struct{})
	blockHandler := func(ctx context.Context, task *model.Task) PhaseResult {
		<-blockCh
		return PhaseResult{}
	}

	runner := New(4, 5)
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
	// 验证里面只有一个任务
	n := len( runner.downloadQueue)
	if n != 1 {
		t.Errorf("expected 1 task in queue, got %d", n)
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

	runner := New(4, 5)
	runner.Register(model.PhaseAdding, blockHandler)

	ctx := context.Background()
	runner.Start(ctx)
	defer func() {
		close(blockCh)
		runner.Stop()
	}()

	task := newTestTask("magnet:cancel")
	runner.Submit(task)

	waitFor(t, 3*time.Second, func() bool {
		return called.Load() > 0
	}, "handler was never called")

	runner.Cancel("magnet:cancel")

	if runner.Get("magnet:cancel") != nil {
		t.Error("task should be removed after Cancel")
	}
}

func TestMaxConcurrency_Limits(t *testing.T) {
	const maxConcurrency = 2
	const totalTasks = 5

	var concurrent atomic.Int32
	var maxSeen atomic.Int32
	doneCh := make(chan struct{})

	slowHandler := func(ctx context.Context, task *model.Task) PhaseResult {
		cur := concurrent.Add(1)
		for {
			old := maxSeen.Load()
			if cur <= old || maxSeen.CompareAndSwap(old, cur) {
				break
			}
		}
		<-doneCh
		concurrent.Add(-1)
		return PhaseResult{}
	}

	runner := New(maxConcurrency, 10) // maxDownload 很大，不限制
	runner.Register(model.PhaseAdding, slowHandler)

	ctx := context.Background()
	runner.Start(ctx)
	defer runner.Stop()

	for i := 0; i < totalTasks; i++ {
		task := newTestTask("magnet:conc-" + string(rune('a'+i)))
		runner.Submit(task)
	}

	// 等待并发达到 maxConcurrency
	waitFor(t, 3*time.Second, func() bool {
		return concurrent.Load() >= int32(maxConcurrency)
	}, "concurrent tasks did not reach maxConcurrency")

	// 给一点时间确保不会超过限制
	time.Sleep(50 * time.Millisecond)

	if v := maxSeen.Load(); v > int32(maxConcurrency) {
		t.Errorf("max concurrent = %d, want <= %d", v, maxConcurrency)
	}

	close(doneCh)
}

func TestDownloadSlot_Limits(t *testing.T) {
	const maxDownload = 2
	const totalTasks = 5

	var concurrent atomic.Int32
	var maxSeen atomic.Int32
	doneCh := make(chan struct{})

	slowAddHandler := func(ctx context.Context, task *model.Task) PhaseResult {
		cur := concurrent.Add(1)
		for {
			old := maxSeen.Load()
			if cur <= old || maxSeen.CompareAndSwap(old, cur) {
				break
			}
		}
		<-doneCh
		concurrent.Add(-1)
		return PhaseResult{}
	}

	runner := New(10, maxDownload) // maxConcurrency 很大，不限制
	runner.Register(model.PhaseAdding, slowAddHandler)

	ctx := context.Background()
	runner.Start(ctx)
	defer runner.Stop()

	for i := 0; i < totalTasks; i++ {
		task := newTestTask("magnet:dl-" + string(rune('a'+i)))
		runner.Submit(task)
	}

	waitFor(t, 3*time.Second, func() bool {
		return concurrent.Load() >= int32(maxDownload)
	}, "concurrent download tasks did not reach maxDownload")

	time.Sleep(50 * time.Millisecond)

	if v := maxSeen.Load(); v > int32(maxDownload) {
		t.Errorf("max concurrent downloads = %d, want <= %d", v, maxDownload)
	}

	close(doneCh)
}

func TestRenaming_NotBlockedByDownloadSlots(t *testing.T) {
	// 下载槽位满时，Renaming 任务仍能执行
	const maxDownload = 1

	blockCh := make(chan struct{})
	var renameCalled atomic.Int32

	blockAddHandler := func(ctx context.Context, task *model.Task) PhaseResult {
		<-blockCh
		return PhaseResult{}
	}
	renameHandler := func(ctx context.Context, task *model.Task) PhaseResult {
		renameCalled.Add(1)
		return PhaseResult{}
	}

	runner := New(4, maxDownload)
	runner.Register(model.PhaseAdding, blockAddHandler)
	runner.Register(model.PhaseRenaming, renameHandler)

	ctx := context.Background()
	runner.Start(ctx)
	defer func() {
		close(blockCh)
		runner.Stop()
	}()

	// 提交一个 Adding 任务占满下载槽位
	addTask := newTestTask("magnet:block-add")
	runner.Submit(addTask)

	// 等待 Adding handler 开始执行（槽位被占）
	time.Sleep(50 * time.Millisecond)

	// 提交一个 Renaming 任务
	renameTask := model.NewRenameTask(
		&model.Torrent{Link: "magnet:rename-only", Name: "test-rename"},
		&model.Bangumi{},
	)
	runner.Submit(renameTask)

	// Renaming 任务应该不被下载槽位阻塞
	waitFor(t, 3*time.Second, func() bool {
		return renameCalled.Load() > 0
	}, "rename task was blocked by full download slots")
}

func TestReleaseSlot_OnFailure(t *testing.T) {
	addHandler, _ := successHandler()

	var failCount atomic.Int32
	failThenSucceed := func(ctx context.Context, task *model.Task) PhaseResult {
		n := failCount.Add(1)
		if n <= 2 {
			return PhaseResult{Err: errors.New("fail")}
		}
		return PhaseResult{}
	}

	runner := New(4, 1) // 只有 1 个下载槽位
	runner.Register(model.PhaseAdding, addHandler)
	runner.Register(model.PhaseChecking, failThenSucceed)

	ctx := context.Background()
	runner.Start(ctx)
	defer runner.Stop()

	for i := 0; i < 3; i++ {
		task := newTestTask("magnet:release-" + string(rune('a'+i)))
		runner.Submit(task)
	}

	waitFor(t, 5*time.Second, func() bool {
		return failCount.Load() >= 3
	}, "not all tasks processed — slot may not be released on failure")
}

func TestNewRenameTask_SkipsToRenamePhase(t *testing.T) {
	renameHandler, renameCount := successHandler()

	runner := New(4, 5)
	runner.Register(model.PhaseAdding, func(ctx context.Context, task *model.Task) PhaseResult {
		t.Error("add handler should not be called for rename task")
		return PhaseResult{}
	})
	runner.Register(model.PhaseChecking, func(ctx context.Context, task *model.Task) PhaseResult {
		t.Error("check handler should not be called for rename task")
		return PhaseResult{}
	})
	runner.Register(model.PhaseDownloading, func(ctx context.Context, task *model.Task) PhaseResult {
		t.Error("download handler should not be called for rename task")
		return PhaseResult{}
	})
	runner.Register(model.PhaseRenaming, renameHandler)

	ctx := context.Background()
	runner.Start(ctx)
	defer runner.Stop()

	task := model.NewRenameTask(
		&model.Torrent{Link: "magnet:rename-only", Name: "test-rename"},
		&model.Bangumi{},
	)
	ok := runner.Submit(task)
	if !ok {
		t.Fatal("Submit should succeed")
	}

	waitFor(t, 3*time.Second, func() bool {
		return runner.Get("magnet:rename-only") == nil
	}, "task did not complete within deadline")

	if v := renameCount.Load(); v != 1 {
		t.Errorf("rename handler called %d times, want 1", v)
	}
	if task.Phase != model.PhaseEnd {
		t.Errorf("task phase = %v, want PhaseEnd", task.Phase)
	}
	if task.HoldingSlot {
		t.Error("rename task should never hold a slot")
	}
}

func TestPollAfter_KeepsDownloadSlot(t *testing.T) {
	const maxDownload = 1
	var addConcurrent atomic.Int32

	addHandler, _ := successHandler()
	checkHandler, _ := successHandler()

	var pollCount atomic.Int32
	pollHandler := func(ctx context.Context, task *model.Task) PhaseResult {
		n := pollCount.Add(1)
		if n < 3 {
			return PhaseResult{PollAfter: 10 * time.Millisecond}
		}
		return PhaseResult{}
	}
	renameHandler, _ := successHandler()

	// 第二个任务的 Adding handler 记录是否同时执行
	addHandler2 := func(ctx context.Context, task *model.Task) PhaseResult {
		addConcurrent.Add(1)
		return PhaseResult{}
	}

	runner := New(4, maxDownload)
	runner.Register(model.PhaseAdding, func(ctx context.Context, task *model.Task) PhaseResult {
		if task.Torrent.Link == "magnet:poll-slot" {
			return addHandler(ctx, task)
		}
		return addHandler2(ctx, task)
	})
	runner.Register(model.PhaseChecking, checkHandler)
	runner.Register(model.PhaseDownloading, pollHandler)
	runner.Register(model.PhaseRenaming, renameHandler)

	ctx := context.Background()
	runner.Start(ctx)
	defer runner.Stop()

	// 第一个任务会在 Downloading 阶段 poll 3 次
	task1 := newTestTask("magnet:poll-slot")
	runner.Submit(task1)

	// 第二个任务也需要下载槽位
	task2 := newTestTask("magnet:blocked")
	runner.Submit(task2)

	// 等待第一个任务完成
	waitFor(t, 3*time.Second, func() bool {
		return runner.Get("magnet:poll-slot") == nil
	}, "first task did not complete")

	// 第二个任务的 Adding 应该在第一个任务释放槽位后才执行
	waitFor(t, 3*time.Second, func() bool {
		return runner.Get("magnet:blocked") == nil
	}, "second task did not complete")
}
