// Package handler 用来处理 event 的各种事件
package subscribe

import (
	"context"
	"log/slog"
	"time"

	"goto-bangumi/internal/database"
	"goto-bangumi/internal/download"
	"goto-bangumi/internal/eventbus"
	"goto-bangumi/internal/model"
)

type CheckService struct {
	bus eventbus.EventBus
}

func (cs *CheckService) Start(ctx context.Context) {
	ch, unsubscribe := eventbus.Subscribe[model.DownloadCheckEvent](cs.bus, ctx, 100)
	defer unsubscribe()

	for event := range ch {
		// 处理下载检查事件
		go func(data model.DownloadCheckEvent) {
			// 先睡10秒，等下载开始
			time.Sleep(10 * time.Second)
			for _, guid := range data.Guids {
				trueID, err := download.Client.Check(ctx, guid)
				if err != nil {
					// 处理错误，例如记录日志
					slog.Error("[check service] 检查下载失败:", "error", err)
				}
				if trueID != "" {
					// 用这个有效的去更新下载状态, 发布下载中事件
					// TODO: 未实现细节
					data.Torrent.DownloadUID = trueID
					// 存入数据库中
					db := database.GetDB()
					if err := db.CreateTorrent(data.Torrent); err != nil {
						slog.Error("[check service] 保存 Torrent 失败:", "error", err)
						return
					}
					// 发布下载中事件
					cs.bus.Publish(ctx, model.DownloadingCheckEvent{
						Torrent: data.Torrent,
						Bangumi: data.Bangumi,
					})
					// 退出循环，检查下一个 guid
					return
				}
			}
		}(event)
	}
}
