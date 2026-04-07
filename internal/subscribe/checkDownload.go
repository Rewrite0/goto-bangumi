// Package subscribe 用来处理 event 的各种事件
package subscribe

import (
	"context"
	"log/slog"
	"time"

	"goto-bangumi/internal/apperrors"
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
	slog.Info("[check service] 下载检查服务已启动")

	for event := range ch {
		// 处理下载检查事件
		go func(data model.DownloadCheckEvent) {
			// 先睡10秒，等下载开始
			time.Sleep(10 * time.Second)
			for _, guid := range data.Guids {
				// 对于网络的错误, 重试3次
				for range 3 {
					trueID, err := download.Client.Check(ctx, guid)
					// 如果表明没有找到，跳出重试循环
					if apperrors.IsKeyError(err) {
						break
					}
					// 如果是登陆错误, 直接返回
					if apperrors.IsDownloadLoginError(err) {
						slog.Error("[check service] 检查下载失败，登录错误:", "error", err)
						return
					}
					if err != nil {
						// 处理错误，例如记录日志
						slog.Error("[check service] 检查下载失败:", "error", err)
					}
					if trueID != "" {
						data.Torrent.DownloadUID = trueID
						// 存入数据库中
						if err := database.GetDB().AddTorrentDUID(ctx, data.Torrent.Link, trueID); err != nil {
							slog.Error("[check service] 更新 Torrent DUID 失败:", "error", err)
							return
						}
						// 发布下载中事件
						cs.bus.Publish(ctx, model.DownloadingCheckEvent{
							Torrent: data.Torrent,
							Bangumi: data.Bangumi,
							Key:     data.Torrent.Link,
						})
						return
					}
				}
			}
		}(event)
	}
}
