// Package scheduler 提供定时任务调度功能
package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Task 定时任务接口
type Task interface {
	// Name 任务名称
	Name() string
	// Interval 执行间隔
	Interval() time.Duration
	// Run 执行任务
	Run(ctx context.Context) error
	// Enable 是否启用
	Enable() bool
}

// Scheduler 调度器
type Scheduler struct {
	tasks  []Task
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

var (
	globalScheduler *Scheduler
	once            sync.Once
)

// NewScheduler 创建新的调度器
func NewScheduler(ctx context.Context) *Scheduler {
	ctx, cancel := context.WithCancel(ctx)
	return &Scheduler{
		tasks:  make([]Task, 0),
		ctx:    ctx,
		cancel: cancel,
	}
}

// GetScheduler 获取全局调度器实例
func GetScheduler() *Scheduler {
	return globalScheduler
}

// InitScheduler 初始化全局调度器
func InitScheduler(ctx context.Context) {
	once.Do(func() {
		globalScheduler = NewScheduler(ctx)
		slog.Info("调度器初始化成功")
	})
}

// AddTask 添加任务
func (s *Scheduler) AddTask(task Task) {
	if task.Enable() {
		s.tasks = append(s.tasks, task)
		slog.Info("添加定时任务", "任务名称", task.Name(), "执行间隔", task.Interval())
	} else {
		slog.Debug("跳过禁用的任务", "任务名称", task.Name())
	}
}

// Start 启动调度器
func (s *Scheduler) Start() {
	if len(s.tasks) == 0 {
		slog.Warn("没有可执行的定时任务")
		return
	}

	slog.Info("启动调度器", "任务数量", len(s.tasks))

	for _, task := range s.tasks {
		s.wg.Add(1)
		go s.runTask(task)
	}
}

// runTask 运行单个任务
func (s *Scheduler) runTask(task Task) {
	defer s.wg.Done()

	ticker := time.NewTicker(task.Interval())
	defer ticker.Stop()

	// 首次立即执行一次
	slog.Debug("执行定时任务", "任务名称", task.Name())
	if err := task.Run(s.ctx); err != nil {
		slog.Error("定时任务执行失败", "任务名称", task.Name(), "错误", err)
	}

	// 定时执行
	for {
		select {
		case <-s.ctx.Done():
			slog.Info("定时任务已停止", "任务名称", task.Name())
			return
		case <-ticker.C:
			slog.Info("执行定时任务", "任务名称", task.Name())
			if err := task.Run(s.ctx); err != nil {
				slog.Error("定时任务执行失败", "任务名称", task.Name(), "错误", err)
			}
		}
	}
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	slog.Info("正在停止调度器...")
	s.cancel()
	s.wg.Wait()
	slog.Info("调度器已停止")
}
