package model

import (
	"context"
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
	CurrentPhase TaskPhase

	HoldingSlot bool            // 是否持有下载槽位
	RetryCount  int             // 当前阶段的重试次数（PollAfter 时自增，advance 时重置）
	Ctx         context.Context // 当前阶段的上下文（如果有）
	CancelFunc  func()          // 取消当前阶段的上下文（如果有）

	// 业务数据
	Guids     []string  // 可能的 hash 列表
	StartTime time.Time // 开始下载时间（用于超时判断）
	NextPoll  time.Time // 下一次轮询时间（用于调度）
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
