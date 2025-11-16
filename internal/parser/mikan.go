package parser

import (
	"bytes"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"goto-bangumi/internal/apperrors"
	"goto-bangumi/internal/model"
	"goto-bangumi/internal/network"
	"goto-bangumi/internal/utils"

	"golang.org/x/net/html"
)

type MikanParser struct{}

func NewMikanParser() *MikanParser {
	return &MikanParser{}
}

func (p *MikanParser) Parse(homepage string) (*model.MikanItem, error) {
	// Fetch HTML content from the URL
	client := network.GetRequestClient()
	content, err := client.Get(homepage)
	if err != nil {
		// network 层已经返回 NetworkError，直接传递
		return nil, err
	}
	// parse mikan html content
	return p.parseHTML(content, homepage)
}

func (p *MikanParser) PosterParse(homepage string) (string, error) {
	// Fetch HTML content from the URL
	client := network.GetRequestClient()
	content, err := client.Get(homepage)
	if err != nil {
		return "", err
	}

	doc, err := html.Parse(bytes.NewReader(content))
	if err != nil {
		return "", &apperrors.ParseError{Err: fmt.Errorf("failed to parse HTML: %w", err)}
	}
	return p.extractPosterLink(doc, homepage)
}

// parseHTML 解析 Mikan 网页的 HTML 内容
func (p *MikanParser) parseHTML(content []byte, pageURL string) (*model.MikanItem, error) {
	doc, err := html.Parse(bytes.NewReader(content))
	if err != nil {
		return nil, &apperrors.ParseError{Err: fmt.Errorf("failed to parse HTML: %w", err)}
	}
	info := model.NewMikanItem()

	// 1. 查找 RSS 链接并提取 Mikan ID
	rssLink := p.findRSSLink(doc)
	if rssLink == "" {
		return nil, &apperrors.ParseError{Err: fmt.Errorf("RSS link not found")}
	}
	id, err := p.extractMikanID(rssLink)
	if err != nil {
		return nil, &apperrors.ParseError{Err: fmt.Errorf("failed to extract Mikan ID: %w", err)}
	}
	info.ID = id

	// 2. 查找 <p class="bangumi-title"> 中的官方标题
	officialTitle := p.findBangumiTitle(doc)
	if officialTitle == "" {
		return info, nil // 找不到标题，返回空信息
	}
	info.OfficialTitle = strings.TrimSpace(officialTitle)
	// 去除标题中的季度信息
	season := MikanSeasonPaattern.FindString(info.OfficialTitle)

	info.Season = 1 // 默认季度为1
	for _, s := range season {
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

	// 3. 提取封面图片链接
	info.PosterLink, err = p.extractPosterLink(doc, pageURL)
	if err != nil {
		return info, err
	}

	return info, nil
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
func (p *MikanParser) extractMikanID(rssURL string) (int, error) {
	// 示例: /RSS/Bangumi?bangumiId=3060&subgroupid=583
	u, err := url.Parse(rssURL)
	if err != nil {
		return 0, &apperrors.ParseError{Err: fmt.Errorf("failed to parse RSS URL: %w", err)}
	}

	bangumiID := u.Query().Get("bangumiId")
	// subgroupID := u.Query().Get("subgroupid")

	if bangumiID == "" {
		return 0, &apperrors.ParseError{Err: fmt.Errorf("bangumiId not found in URL")}
	}

	// 转换为整数返回
	id, err := strconv.Atoi(bangumiID)
	if err != nil {
		return 0, &apperrors.ParseError{Err: fmt.Errorf("invalid bangumiId: %w", err)}
	}
	return id, nil
}

// extractPosterLink 提取封面图片链接
func (p *MikanParser) extractPosterLink(n *html.Node, pageURL string) (string, error) {
	// 查找 <div class="bangumi-poster" style="background-image: url('...')">
	posterDiv := p.findBangumiPoster(n)
	if posterDiv == "" {
		return "", &apperrors.ParseError{Err: fmt.Errorf("bangumi-poster div not found")}
	}

	// 从 style 属性中提取 URL
	// 示例: "background-image: url('/images/Bangumi/...')"
	posterLink := utils.ExtractURLFromStyle(posterDiv)
	if posterLink == "" {
		return "", &apperrors.ParseError{Err: fmt.Errorf("poster URL not found in style")}
	} // 解析根域名
	u, err := url.Parse(pageURL)
	if err != nil {
		return posterLink, &apperrors.ParseError{Err: fmt.Errorf("failed to parse page URL: %w", err)}
	}
	// 拼接完整 URL
	posterLink = strings.Split(posterLink, "?")[0]
	if strings.HasPrefix(posterLink, "/") {
		return fmt.Sprintf("https://%s%s", u.Host, posterLink), nil
	}
	return posterLink, nil
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
