package refresh

import (
	"context"
	"errors"
	"log/slog"

	"gorm.io/gorm"

	"goto-bangumi/internal/database"
	"goto-bangumi/internal/model"
	"goto-bangumi/internal/network"
	"goto-bangumi/internal/taskrunner"
)

// 流程为: 取种子列表 -> 对比数据库中已有的种子 -> 返回新增的种子 -> 检查是否有对应的番剧信息
// -> 如果有 调用 filter, 反回符合条件的种子
// -> 如果没有, 先过一下基础 filter, 然后调用 解析

// Refresher 封装了刷新操作所需的数据库依赖
type Refresher struct {
	db *database.DB
}

// New 创建 Refresher 实例
func New(db *database.DB) *Refresher {
	return &Refresher{db: db}
}

func (r *Refresher) getTorrents(ctx context.Context, url string) []*model.Torrent {
	client := network.GetRequestClient()
	torrents, _ := client.GetTorrents(ctx, url)
	slog.Debug("[getTorrents]从 RSS 获取种子列表", "URL", url, "数量", len(torrents))
	newTorrents, _ := r.db.CheckNewTorrents(ctx, torrents)
	return newTorrents
}

// FindNewBangumi 从 rss 里面看看没有没新的番剧
func (r *Refresher) FindNewBangumi(ctx context.Context, rssItem *model.RSSItem) {
	slog.Info("[FindNewBangumi]检查 RSS 是否有新的番剧", "RSS 名称", rssItem.Name)
	netClient := network.GetRequestClient()
	torrents, _ := netClient.GetTorrents(ctx, rssItem.Link)
	for _, t := range torrents {
		// 突然想起来, possess title 后,名字会和 torrent 里面的差很多,这时就会导致不停的创建
		// 这就是之前 AB 会导致不停的创建的原因, 新在已经解决了
		// 解决方案是对 torrent name 在 get 的时候就处理名字
		_, err := r.db.GetBangumiParseByTitle(ctx, t.Name)
		// 没有找到, 说明是新的番剧
		// 先过一下基础 filter
		if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Debug("[FindNewBangumi]没有找到番剧信息，可能是新的番剧", "种子名称", t.Name, "error", err)
			if FilterTorrent(t, rssItem.ExcludeFilter, rssItem.IncludeFilter) {
				// 要进行一个去重, 一些torrent 是没必要都解析的
				// 进行 metaparser 解析
				slog.Info("[FindNewBangumi]发现新的番剧", "种子名称", t.Name)
				r.createBangumi(ctx, t, rssItem)
			}
		}
	}
}

func (r *Refresher) RefreshRSS(ctx context.Context, url string, runner *taskrunner.TaskRunner) {
	slog.Info("[RefreshRSS]刷新 RSS", "URL", url)
	torrents := r.getTorrents(ctx, url)
	slog.Debug("[RefreshRSS]获取种子列表", "数量", len(torrents))
	for _, t := range torrents {
		metaData, err := r.db.GetBangumiParseByTitle(ctx, t.Name)
		slog.Debug("[RefreshRSS]检查番剧信息", "种子名称", t.Name, "error", err)
		if err != nil {
			continue
		}
		if FilterTorrent(t, metaData.IncludeFilter, metaData.ExcludeFilter) {
			t.Bangumi = metaData
			_ = r.db.CreateTorrent(ctx, t)
			runner.Submit(model.NewAddTask(t, t.Bangumi))
		}
	}
}
