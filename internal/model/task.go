package model

import (
	"context"
	"sync"
	"time"
)

// TaskPhase 任务阶段
type TaskPhase int

const (
	PhaseAdding      TaskPhase = iota // 添加到下载器
	PhaseChecking                     // 检查下载是否成功添加
	PhaseDownloading                  // 下载中，等待完成
	PhaseRenaming                     // 重命名文件
	PhaseCompleted                    // 完成
	PhaseFailed                       // 失败
	PhaseEnd                          // 任务完成标志
)

type TaskState int

const (
	TaskStateCreated   TaskState = iota // 已创建，尚未执行
	TaskStateReady                      // 当前可以执行，等待 scheduler 调度
	TaskStateWaiting                    // 等待 PollAfter 到期
	TaskStateQueued                     // 已被 scheduler 选中，等待 worker 开始
	TaskStateRunning                    // worker 正在执行 handler
	TaskStateCompleted                  // 已结束，不再调度
)

func (p TaskPhase) String() string {
	switch p {
	case PhaseAdding:
		return "adding"
	case PhaseChecking:
		return "checking"
	case PhaseDownloading:
		return "downloading"
	case PhaseRenaming:
		return "renaming"
	case PhaseCompleted:
		return "completed"
	case PhaseFailed:
		return "failed"
	case PhaseEnd:
		return "end"
	default:
		return "unknown"
	}
}

// IsTerminal 是否为终态
func (p TaskPhase) IsTerminal() bool {
	return p == PhaseEnd
}

// Task 下载任务
type Task struct {
	sync.Mutex

	CurrentPhase TaskPhase
	State        TaskState

	RetryCount int             // 当前阶段的可重试错误次数，由 handler 维护，advance 时重置
	Ctx        context.Context // 当前阶段的上下文（如果有）
	CancelFunc func()          // 取消当前阶段的上下文（如果有）

	// 业务数据
	Guids     []string  // 可能的 hash 列表
	StartTime time.Time // 开始下载时间（用于超时判断）
	NextPoll  time.Time // Waiting 状态的预计唤醒时间
	EndTime   time.Time // 结束时间（成功或失败）
	ErrorMsg  string

	// 关联对象（内存引用）
	Torrent *Torrent
	Bangumi *Bangumi
}

// NewAddTask 创建下载任务（从 PhaseAdding 开始）
func NewAddTask(torrent *Torrent, bangumi *Bangumi) *Task {
	return &Task{
		CurrentPhase: PhaseAdding,
		State:        TaskStateCreated,
		Torrent:      torrent,
		Bangumi:      bangumi,
	}
}

// NewRenameTask 创建重命名任务（从 PhaseRenaming 开始）
func NewRenameTask(torrent *Torrent, bangumi *Bangumi) *Task {
	return &Task{
		CurrentPhase: PhaseRenaming,
		Torrent:      torrent,
		Bangumi:      bangumi,
	}
}
