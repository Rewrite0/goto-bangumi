package taskrunner

import (
	"sync"

	"goto-bangumi/internal/model"
)

// type taskQueue struct {
// 	mu       sync.Mutex
// 	download []*model.Task
// 	general  []*model.Task
// }

type taskQueue struct{
	mu sync.Mutex
	queue []*model.Task
}

// enqueue 将任务添加到队列末尾
func (q *taskQueue) enqueue(task *model.Task) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.queue = append(q.queue, task)
	// if needsDownloadSlot(task.CurrentPhase) && (task.HoldingSlot || task.StartTime.IsZero()) {
	// 	q.download = append(q.download, task)
	// } else {
	// 	q.general = append(q.general, task)
	// }
}


// remove 从队列中删除指定任务（指针比较），任务只在一个队列里所以找到即返回
func (q *taskQueue) remove(task *model.Task) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for i, t := range q.queue {
		if t == task {
			q.queue = append(q.queue[:i], q.queue[i+1:]...)
			return
		}
	}
}

// tryDequeue 原子地查找并移除一个可调度的任务，无任务时返回 nil
// 优先已持有槽位的任务，其次在槽位未满时取队首
// func (q *taskQueue) tryDequeue() *model.Task {
// 	q.mu.Lock()
// 	defer q.mu.Unlock()
// 	idx := -1
// 	for i, task := range q.queue {
// 		if task.HoldingSlot {
// 			idx = i
// 			break
// 		}
// 	}
// 	if idx < 0 && downloadSlots < maxDownload && len(q.queue) > 0 {
// 		idx = 0
// 	}
// 	if idx < 0 {
// 		return nil
// 	}
// 	task := q.queue[idx]
// 	q.queue = append(q.queue[:idx], q.queue[idx+1:]...)
// 	return task
// }

// tryDequeueGeneral 移除并返回 general 队列队首任务，队列为空时返回 nil
func (q *taskQueue) tryDequeue() *model.Task {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.queue) == 0 {
		return nil
	}
	task := q.queue[0]
	q.queue = q.queue[1:]
	return task
}

// lenDownload 返回 download 队列长度
// func (q *taskQueue) lenDownload() int {
// 	q.mu.Lock()
// 	defer q.mu.Unlock()
// 	return len(q.download)
// }
