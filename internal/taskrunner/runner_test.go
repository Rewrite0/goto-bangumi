package taskrunner

import (
	"context"
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
