package utils

import (
	htmlpkg "html"
	"slices"
	"strings"

	"golang.org/x/net/html"
)

// HasClass 检查节点是否包含指定的 class
func HasClass(n *html.Node, className string) bool {
	for _, attr := range n.Attr {
		if attr.Key == "class" {
			classes := strings.Fields(attr.Val)
			return slices.Contains(classes, className)
		}
	}
	return false
}

// GetAttr 获取节点的属性值
func GetAttr(n *html.Node, attrName string) string {
	for _, attr := range n.Attr {
		if attr.Key == attrName {
			return attr.Val
		}
	}
	return ""
}

// GetTextContent 获取节点的文本内容
func GetTextContent(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var result strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result.WriteString(GetTextContent(c))
	}
	return result.String()
}

// ExtractURLFromStyle 从 style 属性中提取 URL
// 例如: "background-image: url('/images/Bangumi/...?width=400&amp;height=560')"
func ExtractURLFromStyle(style string) string {
	// 查找 "url(" 开始位置
	prefix := "url("
	start := strings.Index(style, prefix)
	if start == -1 {
		return ""
	}

	// 跳过 "url("
	content := style[start+len(prefix):]

	// 提取引号内的 URL
	var urlPart string
	if strings.HasPrefix(content, "'") {
		// 情况: url('...')
		before, _, found := strings.Cut(content[1:], "'")
		if !found {
			return ""
		}
		urlPart = before
	} else if strings.HasPrefix(content, `"`) {
		// 情况: url("...")
		before, _, found := strings.Cut(content[1:], `"`)
		if !found {
			return ""
		}
		urlPart = before
	} else {
		// 情况: url(...)
		before, _, found := strings.Cut(content, ")")
		if !found {
			return ""
		}
		urlPart = before
	}

	// 解码 HTML 实体 &amp; -> &
	return htmlpkg.UnescapeString(urlPart)
}
