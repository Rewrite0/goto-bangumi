package utils

import (
	"strings"
)
func ProcessTitle(title string) string {
	// title 里面可能有"\n"
	title = strings.ReplaceAll(title, "\n", "")
	// 如果以【开头
	if strings.HasPrefix(title, "【") {
		title = strings.ReplaceAll(title, "【", "[")
		title = strings.ReplaceAll(title, "】", "]")
	}
	title = strings.TrimSpace(title)
	return title
}
