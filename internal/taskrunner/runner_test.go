package taskrunner

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"goto-bangumi/internal/model"
)

func TestSchedulePollsHoldingDownloadTasksWhenSlotsAreFull(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runner := New(4, 2)
	var calls atomic.Int32
	runner.Register(model.PhaseAdding, func(ctx context.Context, task *model.Task) PhaseResult {
		calls.Add(1)
		return PhaseResult{PollAfter: 10 * time.Millisecond}
	})
	runner.Start(ctx)
	defer runner.Stop()

	for i := range 2 {
		runner.Submit(model.NewAddTask(
			&model.Torrent{Link: fmt.Sprintf("torrent-%d", i), Name: fmt.Sprintf("torrent %d", i)},
			model.NewBangumi(),
		))
	}

	waitUntil(t, time.Second, func() bool {
		return calls.Load() >= 4
	})

	for i := range 2 {
		runner.Cancel(fmt.Sprintf("torrent-%d", i))
	}
}

func TestCancelReleasesDownloadSlot(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runner := New(2, 1)
	runner.Register(model.PhaseAdding, func(ctx context.Context, task *model.Task) PhaseResult {
		return PhaseResult{PollAfter: time.Hour}
	})
	runner.Start(ctx)
	defer runner.Stop()

	runner.Submit(model.NewAddTask(
		&model.Torrent{Link: "torrent", Name: "torrent"},
		model.NewBangumi(),
	))

	waitUntil(t, time.Second, func() bool {
		runner.mu.Lock()
		defer runner.mu.Unlock()
		return len(runner.downloadSlots) == 1
	})

	runner.Cancel("torrent")

	waitUntil(t, time.Second, func() bool {
		runner.mu.Lock()
		defer runner.mu.Unlock()
		return len(runner.downloadSlots) == 0
	})
}

func TestTaskStateTracksWorkerLifecycle(t *testing.T) {
	runner := New(1, 1)
	started := make(chan struct{})
	finish := make(chan struct{})
	runner.Register(model.PhaseAdding, func(ctx context.Context, task *model.Task) PhaseResult {
		close(started)
		<-finish
		return PhaseResult{}
	})

	task := model.NewAddTask(
		&model.Torrent{Link: "torrent", Name: "torrent"},
		model.NewBangumi(),
	)
	if !runner.Submit(task) {
		t.Fatal("submit task failed")
	}

	runner.runningSem <- struct{}{}
	if picked := runner.pickTask(); picked != task {
		t.Fatalf("picked task = %p, want %p", picked, task)
	}
	if state := taskState(task); state != model.TaskStateQueued {
		t.Fatalf("state after pick = %v, want queued", state)
	}
	if picked := runner.pickTask(); picked != nil {
		t.Fatal("queued task was picked again")
	}

	runner.dispatch(task)
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("handler did not start")
	}
	if state := taskState(task); state != model.TaskStateRunning {
		t.Fatalf("state in handler = %v, want running", state)
	}
	if picked := runner.pickTask(); picked != nil {
		t.Fatal("running task was picked again")
	}

	close(finish)
	runner.wg.Wait()
	if state := taskState(task); state != model.TaskStateReady {
		t.Fatalf("state after worker exit = %v, want ready", state)
	}
	if task.CurrentPhase != model.PhaseChecking {
		t.Fatalf("phase after worker exit = %v, want checking", task.CurrentPhase)
	}
	runner.Cancel(task.Torrent.Link)
}

func TestPollAfterTransitionsWaitingTaskToReady(t *testing.T) {
	runner := New(1, 1)
	runner.Register(model.PhaseAdding, func(ctx context.Context, task *model.Task) PhaseResult {
		return PhaseResult{PollAfter: 100 * time.Millisecond}
	})

	task := model.NewAddTask(
		&model.Torrent{Link: "torrent", Name: "torrent"},
		model.NewBangumi(),
	)
	task.RetryCount = 3
	if !runner.Submit(task) {
		t.Fatal("submit task failed")
	}
	runner.runningSem <- struct{}{}
	if picked := runner.pickTask(); picked != task {
		t.Fatalf("picked task = %p, want %p", picked, task)
	}
	runner.dispatch(task)
	runner.wg.Wait()

	if state := taskState(task); state != model.TaskStateWaiting {
		t.Fatalf("state after PollAfter = %v, want waiting", state)
	}
	if task.RetryCount != 3 {
		t.Fatalf("retry count after PollAfter = %d, want 3", task.RetryCount)
	}
	if picked := runner.pickTask(); picked != nil {
		t.Fatal("waiting task was picked before PollAfter elapsed")
	}

	waitUntil(t, time.Second, func() bool {
		return taskState(task) == model.TaskStateReady
	})
	if picked := runner.pickTask(); picked != task {
		t.Fatalf("picked task = %p, want ready task %p", picked, task)
	}
	runner.Cancel(task.Torrent.Link)
}

func TestOldWorkerDoesNotRemoveResubmittedTask(t *testing.T) {
	runner := New(1, 1)
	started := make(chan struct{})
	finish := make(chan struct{})
	oldTask := model.NewAddTask(
		&model.Torrent{Link: "torrent", Name: "old torrent"},
		model.NewBangumi(),
	)
	runner.Register(model.PhaseAdding, func(ctx context.Context, task *model.Task) PhaseResult {
		if task == oldTask {
			close(started)
			<-finish
		}
		return PhaseResult{Err: errors.New("handler stopped")}
	})

	if !runner.Submit(oldTask) {
		t.Fatal("submit old task failed")
	}
	runner.runningSem <- struct{}{}
	if picked := runner.pickTask(); picked != oldTask {
		t.Fatalf("picked task = %p, want old task %p", picked, oldTask)
	}
	runner.dispatch(oldTask)
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("old handler did not start")
	}

	runner.Cancel(oldTask.Torrent.Link)
	newTask := model.NewAddTask(
		&model.Torrent{Link: "torrent", Name: "new torrent"},
		model.NewBangumi(),
	)
	if !runner.Submit(newTask) {
		t.Fatal("submit replacement task failed")
	}
	if picked := runner.pickTask(); picked != newTask {
		t.Fatalf("picked task = %p, want new task %p", picked, newTask)
	}

	close(finish)
	runner.wg.Wait()

	runner.mu.Lock()
	gotTask := runner.tasks[newTask.Torrent.Link]
	gotSlot := runner.downloadSlots[newTask.Torrent.Link]
	runner.mu.Unlock()
	if gotTask != newTask {
		t.Fatalf("task after old worker exit = %p, want new task %p", gotTask, newTask)
	}
	if gotSlot != newTask {
		t.Fatalf("slot after old worker exit = %p, want new task %p", gotSlot, newTask)
	}
	runner.Cancel(newTask.Torrent.Link)
}

func taskState(task *model.Task) model.TaskState {
	task.Lock()
	defer task.Unlock()
	return task.State
}

func waitUntil(t *testing.T, timeout time.Duration, ok func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if ok() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("condition not met before timeout")
}
