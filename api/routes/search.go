package routes

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"

	"goto-bangumi/api/response"
)

// SearchResult 搜索结果
type SearchResult struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	Size        string `json:"size"`
	Seeders     int    `json:"seeders"`
	Leechers    int    `json:"leechers"`
	PublishDate string `json:"publish_date"`
	Provider    string `json:"provider"`
}

// SearchProvider 搜索提供商
type SearchProvider struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

// RegisterSearchRoutes 注册搜索路由
func RegisterSearchRoutes(r *gin.RouterGroup) {
	search := r.Group("/search")
	{
		search.GET("/bangumi", searchBangumi)
		search.GET("/provider", getProviders)
	}
}

// searchBangumi 搜索番剧
// GET /api/v1/search/bangumi?keyword=xxx&site=xxx
// 使用 SSE (Server-Sent Events) 实时返回搜索结果
func searchBangumi(c *gin.Context) {
	keyword := c.Query("keyword")
	site := c.Query("site")

	if len(keyword) < 2 {
		response.BadRequest(c, "Keyword must be at least 2 characters", "关键词至少需要2个字符")
		return
	}

	// 设置 SSE 响应头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	// 获取客户端通知通道
	clientGone := c.Request.Context().Done()

	// TODO: 实现真正的搜索逻辑
	// 这里需要调用 searcher 模块进行搜索
	// searcher := search.NewSearcher(site)
	// results := searcher.Search(keyword)

	// 模拟搜索结果（实际实现时替换）
	go func() {
		defer func() {
			// 发送结束事件
			writeSSE(c.Writer, "done", map[string]string{"message": "Search completed"})
		}()

		// TODO: 从 searcher 获取结果并发送
		// for result := range results {
		//     select {
		//     case <-clientGone:
		//         return
		//     default:
		//         writeSSE(c.Writer, "result", result)
		//     }
		// }

		_ = keyword // 消除未使用警告
		_ = site
		_ = clientGone
	}()

	// 阻塞直到客户端断开
	<-clientGone
}

// writeSSE 写入 SSE 事件
func writeSSE(w io.Writer, event string, data any) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}

	fmt.Fprintf(w, "event: %s\n", event)
	fmt.Fprintf(w, "data: %s\n\n", jsonData)

	if f, ok := w.(interface{ Flush() }); ok {
		f.Flush()
	}
}

// getProviders 获取搜索提供商列表
// GET /api/v1/search/provider
func getProviders(c *gin.Context) {
	// TODO: 从配置或 searcher 模块获取实际的提供商列表

	providers := []SearchProvider{
		{
			ID:      "mikan",
			Name:    "Mikan Project",
			Enabled: true,
		},
		{
			ID:      "nyaa",
			Name:    "Nyaa",
			Enabled: true,
		},
		{
			ID:      "dmhy",
			Name:    "动漫花园",
			Enabled: true,
		},
		{
			ID:      "acgrip",
			Name:    "ACG.RIP",
			Enabled: true,
		},
	}

	response.Success(c, providers)
}
