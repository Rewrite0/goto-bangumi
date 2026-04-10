package subscribe

import (
	"context"
	"log/slog"

	"goto-bangumi/internal/database"
	"goto-bangumi/internal/eventbus"
	"goto-bangumi/internal/model"
	"goto-bangumi/internal/rename"
)

// renameService 处理重命名事件
type renameService struct {
	bus    eventbus.EventBus
	db     *database.DB
	renamer *rename.Renamer
}

// handleRename 处理单个重命名事件
func (rs *renameService) handleRename(ctx context.Context, data model.RenameEvent) {
	slog.Info("[rename service] 收到重命名事件", "torrent", data.Torrent.Name, "bangumi", data.Bangumi.OfficialTitle)

	// 调用 rename 模块进行重命名
	rs.renamer.Rename(ctx, data.Torrent, data.Bangumi)

	// 更新数据库状态为已重命名
	if err := rs.db.TorrentRenamed(ctx, data.Torrent.Link); err != nil {
		slog.Error("[rename service] 更新种子重命名状态失败", "error", err, "link", data.Torrent.Link)
		return
	}

	slog.Info("[rename service] 重命名完成", "torrent", data.Torrent.Name)
}

// Start 启动重命名服务
func (rs *renameService) Start(ctx context.Context) {
	ch, unsubscribe := eventbus.Subscribe[model.RenameEvent](rs.bus, ctx, 100)
	defer unsubscribe()
	slog.Info("[rename service] 重命名服务已启动")

	for event := range ch {
		go rs.handleRename(ctx, event)
	}
}
