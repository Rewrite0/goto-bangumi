package core

import (
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
)


// 先实现一下整体的初使化

func InitProgram()  {
	// database
	// 开始 event bus
	err := conf.LoadConfig()
	if err != nil {
		panic(err)
	}

	// 获取程序配置并初始化日志
	programConfig := conf.GetConfigOrDefault("program", model.NewProgramConfig())
	logger.Init(programConfig.DebugEnable)
	// 初始化数据库
	if err:=database.InitDB(nil); err != nil {
		slog.Error("初始化数据库失败", "error", err)
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
		fmt.Println("配置文件已更新")
	}

}
