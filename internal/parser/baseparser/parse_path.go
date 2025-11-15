package baseparser

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strconv"
	"strings"
)

// PathInfo 路径解析结果
type PathInfo struct {
	BangumiName  string // 番剧名称
	Year         string // 年份（可能为空）
	SeasonNumber int    // 季度编号
}

// ParsePath 从路径中解析 bangumi name, season number 和 year
// 输入格式示例:
//   - "进击的巨人 (2013)/Season 1"
//   - "进击的巨人/Season 2"
//   - "Frieren (2023)/Season 01"
func ParsePath(relativePath string) *PathInfo {
	// 从中拿到 bangumi name,season, year
	parts := strings.Split(relativePath, string(filepath.Separator))
	if len(parts) != 2 {
		slog.Error("[parser] Invalid relative path format", "relativePath", relativePath)
		return nil
	}

	bangumiPart := parts[0] // BangumiName (Year)
	seasonPart := parts[1]  // Season \d

	info := &PathInfo{
		BangumiName:  "",
		Year:         "",
		SeasonNumber: 1, // 默认值
	}

	// 1. 从 bangumiPart 提取 bangumi name 和 year（可选）
	info.BangumiName, info.Year = parseBangumiPart(bangumiPart)

	// 2. 从 seasonPart 提取 season number
	season, err := parseSeasonPart(seasonPart)
	if err != nil {
		slog.Error("[parser] Failed to parse season part", "seasonPart", seasonPart, "error", err)
		return nil
	}
	info.SeasonNumber = season

	return info
}

// parseBangumiPart 从 bangumiPart 中提取 bangumi name 和 year
// 例如: "进击的巨人 (2013)" -> name: "进击的巨人", year: "2013"
//
//	"进击的巨人" -> name: "进击的巨人", year: ""
func parseBangumiPart(bangumiPart string) (name, year string) {
	name = bangumiPart
	year = ""

	// 检查是否有括号包裹的年份
	if !strings.HasSuffix(bangumiPart, ")") {
		return strings.TrimSpace(name), year
	}

	// 找到最后一个 "(" 的位置
	idx := strings.LastIndex(bangumiPart, "(")
	if idx == -1 {
		return strings.TrimSpace(name), year
	}

	// 提取括号内的内容
	possibleYear := strings.TrimSpace(bangumiPart[idx+1 : len(bangumiPart)-1])

	// 检查是否是4位数字
	if len(possibleYear) == 4 {
		if _, err := strconv.Atoi(possibleYear); err == nil {
			// 确实是年份
			year = possibleYear
			name = strings.TrimSpace(bangumiPart[:idx])
		}
	}

	return name, year
}

// parseSeasonPart 从 seasonPart 中提取 season number
// 例如: "Season 1" -> 1
//
//	"Season 01" -> 1
//	"season 2" -> 2
func parseSeasonPart(seasonPart string) (int,error) {
	// 移除 "season" 前缀（不区分大小写）
	seasonStrs := strings.Split(seasonPart, " ")
	seasonStr := seasonStrs[len(seasonStrs)-1]
	seasonStr = strings.TrimSpace(seasonStr)

	// 尝试转换为数字
	if num, err := strconv.Atoi(seasonStr); err == nil {
		return num,nil
	}

	// 解析失败，返回默认值
	return 1, fmt.Errorf("invalid season format: %s", seasonPart)
}
