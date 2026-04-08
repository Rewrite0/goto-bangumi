package core

import (
	"context"
	"log/slog"

	"goto-bangumi/internal/conf"
	"goto-bangumi/internal/database"
	"goto-bangumi/internal/download"
	"goto-bangumi/internal/logger"
	"goto-bangumi/internal/model"
	"goto-bangumi/internal/network"
	"goto-bangumi/internal/notification"
	"goto-bangumi/internal/parser"
	"goto-bangumi/internal/rename"
	"goto-bangumi/internal/scheduler"
	"goto-bangumi/internal/task"
	"goto-bangumi/internal/taskrunner"
	"goto-bangumi/internal/taskrunner/handlers"
)

// 先实现一下整体的初使化

type Program struct {
	// 这里可以添加程序的全局状态和配置
	ctx    context.Context
	cancel context.CancelFunc
}

func InitProgram(ctx context.Context) {
	// Load config
	if err := conf.Init(); err != nil {
		slog.Error("[program] 加载配置文件失败", "error", err)
		panic(err)
	}

	cfg := conf.Get()

	// Initialize logger
	logger.Init(cfg.Program.DebugEnable)

	// Initialize database
	if err := database.InitDB(nil); err != nil {
		slog.Error("[program] 初始化数据库失败", "error", err)
		panic(err)
	}

	// Initialize modules with injected config
	network.Init(&cfg.Proxy)
	parser.Init(&cfg.Parser)
	notification.NotificationClient.Init(&cfg.Notification)
	download.Client.Init(&cfg.Downloader)
	rename.Init(&cfg.Rename)
}

func (p *Program) Start(ctx context.Context) {
	p.ctx, p.cancel = context.WithCancel(ctx)
	go download.Client.Login(p.ctx)

	// 创建并启动 taskrunner
	runner := taskrunner.New(taskrunner.DefaultConfig())
	runner.Register(model.PhaseAdding, handlers.NewAddHandler(), true)          // 唯一受限阶段
	runner.Register(model.PhaseChecking, handlers.NewCheckHandler(), false)     // 轻量查询
	runner.Register(model.PhaseDownloading, handlers.NewDownloadingHandler(), false) // 轻量轮询
	runner.Register(model.PhaseRenaming, handlers.NewRenameHandler(), false)    // 本地文件操作
	runner.Start(p.ctx)

	// 启动调度器
	InitScheduler(p.ctx, runner)
}

func (p *Program) Stop() {
	p.cancel()
	slog.Info("程序已停止")
}

// InitScheduler 初始化并启动调度器
func InitScheduler(ctx context.Context, runner *taskrunner.TaskRunner) {
	scheduler.InitScheduler(ctx)

	s := scheduler.GetScheduler()
	if s == nil {
		slog.Error("调度器初始化失败")
		return
	}

	s.AddTask(task.NewRSSRefreshTask(conf.Get().Program))
	s.AddTask(task.NewDownloadTask(runner))

	s.Start()

	slog.Info("调度器启动成功")
}

//TODO: 日志更新的时候要知道是哪一部分更新了,然后要对哪一部分进行重新初始化
