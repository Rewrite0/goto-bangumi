package model

import (
	"sync"
	"time"
)

// TaskPhase 任务阶段
type TaskPhase int

const (
	PhaseChecking    TaskPhase = iota // 检查下载是否成功添加
	PhaseDownloading                  // 下载中，等待完成
	PhaseRenaming                     // 重命名文件
	PhaseCompleted                    // 完成
	PhaseFailed                       // 失败
	PhaseEnd                          // 任务完成标志
)

func (p TaskPhase) String() string {
	switch p {
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
	Mu    sync.Mutex
	Phase TaskPhase

	// 业务数据
	Guids     []string  // 可能的 hash 列表
	StartTime time.Time // 开始下载时间（用于超时判断）
	ErrorMsg  string

	// 关联对象（内存引用）
	Torrent *Torrent
	Bangumi *Bangumi
}

func NewTask(torrent *Torrent, bangumi *Bangumi) *Task {
	return &Task{
		Phase:     PhaseChecking,
		StartTime: time.Now(),
		Torrent:   torrent,
		Bangumi:   bangumi,
	}
}
