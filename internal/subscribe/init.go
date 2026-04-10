package subscribe

import (
	"context"
	"goto-bangumi/internal/database"
	"goto-bangumi/internal/eventbus"
	"goto-bangumi/internal/rename"
)

func InitModule(eventBus eventbus.EventBus, ctx context.Context, db *database.DB, renamer *rename.Renamer) {
	// 注册检查下载完成的订阅者
	checkService := &CheckService{
		bus: eventBus,
		db:  db,
	}
	go checkService.Start(ctx)

	// 注册检查下载中的订阅者
	checkDownloadingService := &checkDownloadingService{
		bus: eventBus,
		db:  db,
	}
	go checkDownloadingService.Start(ctx)

	// 注册重命名服务的订阅者
	renameService := &renameService{
		bus:     eventBus,
		db:      db,
		renamer: renamer,
	}
	go renameService.Start(ctx)
}
