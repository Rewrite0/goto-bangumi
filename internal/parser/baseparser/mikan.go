package baseparser

import (
	"bytes"
	"fmt"
	"goto-bangumi/internal/network"
	"goto-bangumi/internal/utils"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

// MikanInfo 对应 Python 版本的 MikanInfo 模型
type MikanInfo struct {
	ID            string // Mikan ID (格式: "bangumiId" 或 "bangumiId#subgroupId")
	OfficialTitle string // 官方标题
	Season        int    // 季度
	PosterLink    string // 海报链接
}

type MikanParser struct{}

func NewMikanParser() *MikanParser {
	return &MikanParser{}
}

func (p *MikanParser) Parse(homepage string) (*MikanInfo, error) {
	// Fetch HTML content from the URL
	client, err := network.NewRequestClient()
	if err != nil {
		return nil, err
	}
	content, err := client.Get(homepage)
	if err != nil {
		return nil, err
	}
	// parse mikan html content
	return p.parseHTML(content, homepage)
}

func (p *MikanParser) PosterParser(homepage string) (string, error) {
	// Fetch HTML content from the URL
	client, err := network.NewRequestClient()
	if err != nil {
		return "", err
	}
	content, err := client.Get(homepage)
	if err != nil {
		return "", err
	}
	// parse mikan html content

	doc, _ := html.Parse(bytes.NewReader(content))
	//TODO: error handle
	return p.extractPosterLink(doc, homepage), nil
}

// parseHTML 解析 Mikan 网页的 HTML 内容
func (p *MikanParser) parseHTML(content []byte, pageURL string) (*MikanInfo, error) {
	doc, err := html.Parse(bytes.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var info MikanInfo

	// 1. 查找 <p class="bangumi-title"> 中的官方标题
	officialTitle := p.findBangumiTitle(doc)
	if officialTitle == "" {
		return &info, nil // 找不到标题，返回空信息
	}
	info.OfficialTitle = strings.TrimSpace(officialTitle)
	// episodeInfo := metaParser.Parser(info.OfficialTitle)
	// 去除标题中的季度信息
	season := MikanSeasonPaattern.FindString(info.OfficialTitle)
	for _, s := range season{
		if val, ok := ChineseNumberMap[string(s)]; ok {
			info.Season = val
			break
		}
		if val, ok := ChineseNumberUpperMap[string(s)]; ok {
			info.Season = val
			break
		}
	}
	info.OfficialTitle = MikanSeasonPaattern.ReplaceAllString(info.OfficialTitle, "")



	// 2. 查找 RSS 链接并提取 Mikan ID
	rssLink := p.findRSSLink(doc)
	if rssLink != "" {
		info.ID = p.extractMikanID(rssLink)
	}

	// 3. 提取封面图片链接
	info.PosterLink = p.extractPosterLink(doc, pageURL)


	return &info, nil
}

// findBangumiTitle 查找 <p class="bangumi-title"> 中的文本内容
func (p *MikanParser) findBangumiTitle(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "p" {
		if utils.HasClass(n, "bangumi-title") {
			return utils.GetTextContent(n)
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := p.findBangumiTitle(c); result != "" {
			return result
		}
	}
	return ""
}

// findRSSLink 查找 <a href="/RSS/Bangumi?bangumiId=..."> 链接
func (p *MikanParser) findRSSLink(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "a" {
		href := utils.GetAttr(n, "href")
		if strings.HasPrefix(href, "/RSS/Bangumi?bangumiId=") {
			return href
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := p.findRSSLink(c); result != "" {
			return result
		}
	}
	return ""
}

// extractMikanID 从 RSS URL 中提取 Mikan ID（使用 net/url，不用正则）
func (p *MikanParser) extractMikanID(rssURL string) string {
	// 示例: /RSS/Bangumi?bangumiId=3060&subgroupid=583
	u, err := url.Parse(rssURL)
	if err != nil {
		return ""
	}

	bangumiID := u.Query().Get("bangumiId")
	subgroupID := u.Query().Get("subgroupid")

	if bangumiID == "" {
		return ""
	}

	if subgroupID != "" {
		return bangumiID + "#" + subgroupID
	}
	return bangumiID
}

// extractPosterLink 提取封面图片链接
func (p *MikanParser) extractPosterLink(n *html.Node, pageURL string) string {
	// 查找 <div class="bangumi-poster" style="background-image: url('...')">
	posterDiv := p.findBangumiPoster(n)
	if posterDiv == "" {
		return ""
	}

	// 从 style 属性中提取 URL
	// 示例: "background-image: url('/images/Bangumi/...')"
	posterLink := utils.ExtractURLFromStyle(posterDiv)
	if posterLink == "" {
		return ""
	}

	// 解析根域名
	u, err := url.Parse(pageURL)

	if err != nil {
		return posterLink
	}

	// 拼接完整 URL
	posterLink = strings.Split(posterLink, "?")[0]
	if strings.HasPrefix(posterLink, "/") {
		return fmt.Sprintf("https://%s%s", u.Host, posterLink)
	}
	return posterLink
}

// findBangumiPoster 查找 <div class="bangumi-poster"> 的 style 属性
func (p *MikanParser) findBangumiPoster(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "div" {
		if utils.HasClass(n, "bangumi-poster") {
			return utils.GetAttr(n, "style")
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := p.findBangumiPoster(c); result != "" {
			return result
		}
	}
	return ""
}

