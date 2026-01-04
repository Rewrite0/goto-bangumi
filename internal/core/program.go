package core

import (
	"context"
	"fmt"
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
)

// 先实现一下整体的初使化

type Program struct {
	// 这里可以添加程序的全局状态和配置
	ctx    context.Context
	cancel context.CancelFunc
}

func InitProgram(ctx context.Context) {
	// database
	// 开始 event bus
	conf.Init()
	err := conf.LoadConfig()
	// LoadConfig()
	if err != nil {
		slog.Error("[program]加载配置文件失败", "error", err)
		panic(err)
	}

	// 获取程序配置并初始化日志
	programConfig := conf.GetConfigOrDefault("program", model.NewProgramConfig())
	logger.Init(programConfig.DebugEnable)
	// 初始化数据库
	if err := database.InitDB(nil); err != nil {
		slog.Error("[program]初始化数据库失败", "error", err)
		panic(err)
	}

	// 初始化网络模块
	network.Init(network.GetConfig())

	// 初始化解析器模块
	parser.InitModule()

	// 初始化通知模块
	notification.InitModule()
	// 初始化下载客户端
	// download.Client.Init(download.Client.GetConfig())
	download.InitModule()
	// 初始化重命名模块
	rename.InitModule()

	// 检查配置文件是否需要更新
	if conf.NeedUpdate {
		err := conf.SaveConfig()
		if err != nil {
			panic(err)
		}
		fmt.Println("[program]配置文件已更新")
	}
}

func (p *Program) Start(ctx context.Context) {
	p.ctx, p.cancel = context.WithCancel(ctx)
	go download.Client.Login(p.ctx)
	// 启动调度器
	InitScheduler(p.ctx)
	// 注册事件监听器
}

func (p *Program) Stop() {
	p.cancel()
	slog.Info("程序已停止")
}

// InitScheduler 初始化并启动调度器
// ctx: 上下文，用于控制调度器的生命周期
func InitScheduler(ctx context.Context) {
	// 初始化调度器
	scheduler.InitScheduler(ctx)

	s := scheduler.GetScheduler()
	if s == nil {
		slog.Error("调度器初始化失败")
		return
	}

	// 添加 RSS 刷新任务
	s.AddTask(task.NewRSSRefreshTask())
	s.AddTask(task.NewDownloadTask())

	// 启动调度器
	s.Start()

	slog.Info("调度器启动成功")
}

//TODO: 日志更新的时候要知道是哪一部分更新了,然后要对哪一部分进行重新初始化
