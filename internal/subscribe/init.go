package subscribe

import (
	"context"
	"goto-bangumi/internal/eventbus"
)

func InitModule(eventBus eventbus.EventBus, ctx context.Context) {
	// 注册检查下载完成的订阅者
	checkService := &CheckService{
		bus: eventBus,
	}
	go checkService.Start(ctx)

	// 注册检查下载中的订阅者
	checkDownloadingService := &checkDownloadingService{
		bus: eventBus,
	}
	go checkDownloadingService.Start(ctx)

	// 注册重命名服务的订阅者
	renameService := &renameService{
		bus: eventBus,
	}
	go renameService.Start(ctx)
}
