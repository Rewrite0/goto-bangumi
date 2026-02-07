package taskrunner

import (
	"sync"

	"goto-bangumi/internal/model"
)

// TaskStore 内存任务存储，以 torrent link 为唯一标识
type TaskStore struct {
	mu    sync.Mutex
	tasks map[string]*model.Task // key: torrent link
}

// NewTaskStore 创建任务存储
func NewTaskStore() *TaskStore {
	return &TaskStore{
		tasks: make(map[string]*model.Task),
	}
}

// Add 添加任务。如果 link 已存在则忽略，返回 false
func (s *TaskStore) Add( task *model.Task) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	link := task.Torrent.Link
	if _, exists := s.tasks[link]; exists {
		return false
	}
	// 构建 cancel 用以单独取消任务
	s.tasks[link] = task
	return true
}

// Get 根据 link 获取任务
func (s *TaskStore) Get(link string) *model.Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.tasks[link]
}


// Remove 根据 link 移除任务
func (s *TaskStore) Remove(link string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tasks, link)
}






