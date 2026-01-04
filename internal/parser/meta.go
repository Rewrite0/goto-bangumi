// Package parser 包含对标题的基本解析功能, 额外提供 tmdb 和 mikan 的解析功能, TODO:bangumi解析未做
package parser

import (
	"strconv"
	"strings"

	"goto-bangumi/internal/model"
	"goto-bangumi/internal/parser/patterns"
	"goto-bangumi/internal/utils"

	"github.com/dlclark/regexp2"
)

// Episode 是 model.Episode 的类型别名
// type Episode = model.Episode

// TitleMetaParser 原始视频标题解析器,差不多一秒解析6000个的样子
type TitleMetaParser struct {
	rawTitle       string
	title          string
	token          []string
	episodeTrusted bool
	seasonTrusted  bool
}

// NewTitleMetaParse 创建新的解析器实例
func NewTitleMetaParse() *TitleMetaParser {
	return &TitleMetaParser{
		rawTitle:       "",
		title:          "",
		token:          make([]string, 0),
		episodeTrusted: false,
		seasonTrusted:  false,
	}
}

// findallSubTitle 查找并替换标题中的模式
// 模拟 Python 的 re.findall 行为：返回所有匹配的捕获组
// replace 参数控制是否替换匹配的文本，默认为 true
func (p *TitleMetaParser) findallSubTitle(pattern *regexp2.Regexp, sym string, replace ...bool) [][]string {
	shouldReplace := true
	if len(replace) > 0 {
		shouldReplace = replace[0]
	}

	ans := make([][]string, 0)
	positions := make([][2]int, 0) // 记录匹配位置 [rune index]

	// 使用 FindNextMatch 获取所有匹配
	match, _ := pattern.FindStringMatch(p.title)
	for match != nil {
		// 收集所有捕获组（跳过第0个，即完整匹配）
		groups := make([]string, 0)
		for _, group := range match.Groups()[1:] {
			groups = append(groups, group.String())
		}
		ans = append(ans, groups)

		// 记录整个匹配的位置（regexp2 使用 rune index）
		positions = append(positions, [2]int{match.Index, match.Index + match.Length})

		match, _ = pattern.FindNextMatch(match)
	}

	// 如果找到匹配且需要替换，从后往前替换（避免位置偏移）
	if shouldReplace && len(positions) > 0 {
		// 将字符串转为 rune 切片以支持 Unicode
		runes := []rune(p.title)
		for i := len(positions) - 1; i >= 0; i-- {
			start := positions[i][0]
			end := positions[i][1]
			// 替换：前部分 + sym + 后部分
			runes = append(runes[:start], append([]rune(sym), runes[end:]...)...)
		}
		p.title = string(runes)
	}

	return ans
}

// getGroupInfo 获取字幕组信息
func (p *TitleMetaParser) getGroupInfo() string {
	groupInfo := p.findallSubTitle(patterns.GroupRe, "[]")
	// 提取第一个捕获组并用& 合并多个字幕组信息
	groups := make([]string, 0)
	for _, match := range groupInfo {
		if len(match) > 0 && match[0] != "" {
			groups = append(groups, match[0])
		}
	}
	return strings.TrimSpace(strings.Join(groups, "&"))
}

// getCollectionInfo 获取合集信息
// 返回: isCollection 是否为合集, start 起始集数, end 结束集数
func (p *TitleMetaParser) getCollectionInfo() (isCollection bool, start int, end int) {
	// 1. 先尝试有范围的规则（按优先级顺序）
	for _, pattern := range patterns.CollectionRangePatterns {
		matches := p.findallSubTitle(pattern, "/[]", false)
		for _, match := range matches {
			// 提取前两个非空数字
			var nums []int
			for _, m := range match {
				if m != "" {
					num, err := strconv.Atoi(m)
					if err == nil {
						nums = append(nums, num)
					}
					if len(nums) >= 2 {
						break
					}
				}
			}
			// 验证范围：start < end
			if len(nums) >= 2 && nums[0] < nums[1] {
				// 数值合理，执行替换
				p.findallSubTitle(pattern, "/[]")
				p.episodeTrusted = true
				return true, nums[0], nums[1]
			}
		}
	}

	// 2. 尝试无范围的规则（全12话、vol.1 等）
	for _, pattern := range patterns.CollectionSinglePatterns {
		matches := p.findallSubTitle(pattern, "/[]", false)
		if len(matches) > 0 {
			// 执行替换
			p.findallSubTitle(pattern, "/[]")
			p.episodeTrusted = true
			return true, 0, 0
		}
	}

	return false, 0, 0
}

// episodeInfoToEpisode 从剧集信息元组中提取剧集号
func (p *TitleMetaParser) episodeInfoToEpisode(episodeInfo []string) int {
	for _, info := range episodeInfo {
		if info != "" {
			num, err := strconv.Atoi(info)
			if err == nil {
				return num
			}
			// 尝试解析中文数字（1-10）
			if val, ok := patterns.ChineseNumberMap[info]; ok {
				return val
			}
		}
	}
	return 0
}

// parseEpisode 解析剧集号
func (p *TitleMetaParser) parseEpisode(episodeInfo [][]string, episodeIsTrusted bool) int {
	if len(episodeInfo) == 0 {
		// 实在没找到,返回0
		return 0
	}

	if episodeIsTrusted || len(episodeInfo) == 1 {
		// 可信集数 or 长度为1
		// 秉持尽量返回的思想
		return p.episodeInfoToEpisode(episodeInfo[0])
	}

	untrustedEpisodeList := make([]int, 0)
	for _, ep := range episodeInfo {
		untrustedEpisodeList = append(untrustedEpisodeList, p.episodeInfoToEpisode(ep))
	}

	// 所有的集数一致
	if len(untrustedEpisodeList) > 0 {
		allSame := true
		first := untrustedEpisodeList[0]
		for _, x := range untrustedEpisodeList {
			if x != first {
				allSame = false
				break
			}
		}
		if allSame {
			return first
		}

		if len(untrustedEpisodeList) > 1 {
			second := untrustedEpisodeList[1]
			if second != 480 && second != 720 && second != 1080 {
				return second
			}
		}
		return first
	}

	return 0
}

// getTrustedEpisode 获取可信的剧集信息，返回 -1 表示失败
func (p *TitleMetaParser) getTrustedEpisode() int {
	episodeInfo := p.findallSubTitle(patterns.EpisodePatternTrust, "/[]")
	if len(episodeInfo) == 0 {
		episodeInfo = p.findallSubTitle(patterns.EpisodePatternTrustWithBoundary, "/[]")
	}

	if len(episodeInfo) > 0 {
		p.episodeTrusted = true
		return p.parseEpisode(episodeInfo, true)
	}
	return -1
}

// getUntrustedEpisode 获取不可信的剧集信息
func (p *TitleMetaParser) getUntrustedEpisode() int {
	episodeInfo := p.findallSubTitle(patterns.EpisodeReUntrusted, "[]")
	if len(episodeInfo) > 0 {
		return p.parseEpisode(episodeInfo, false)
	}
	return -1
}

// seasonInfoToSeason 从季度信息元组中提取季度号
func (p *TitleMetaParser) seasonInfoToSeason(seasonInfo []string) int {
	// 从元组中找到第一个有效的季度数据
	for _, season := range seasonInfo {
		if strings.Contains(season, "部分") {
			return 1
		}
		if season != "" {
			// 尝试解析数字
			num, err := strconv.Atoi(season)
			if err == nil {
				return num
			}

			// 检查中文数字
			if val, ok := patterns.ChineseNumberMap[season]; ok {
				return val
			}

			// 检查罗马数字
			if val, ok := patterns.RomanNumbers[season]; ok {
				return val
			}
		}
	}
	return 0
}

// parseSeason 解析季度
func (p *TitleMetaParser) parseSeason(seasonInfo [][]string, seasonIsTrusted bool) (int, string) {
	if len(seasonInfo) > 0 {
		seasonList := make([]int, 0)
		for _, s := range seasonInfo {
			seasonList = append(seasonList, p.seasonInfoToSeason(s))
		}

		if seasonIsTrusted {
			if len(seasonInfo) > 0 && len(seasonInfo[0]) > 0 {
				return seasonList[0], string(seasonInfo[0][0])
			}
			return seasonList[0], ""
		} else {
			// 如果是非可信季度信息，返回第一个有效的季度
			if len(seasonInfo[0]) == 1 && seasonList[0] > 1 && seasonList[0] < 5 {
				p.findallSubTitle(patterns.SeasonPatternUntrusted, "[]")
				return seasonList[0], seasonInfo[0][0]
			}
		}
	}

	return 1, ""
}

// getTrustedSeason 获取可信的季度信息
func (p *TitleMetaParser) getTrustedSeason() (int, string) {
	seasonInfo := p.findallSubTitle(patterns.SeasonPatternTruest, "/[]")
	if len(seasonInfo) == 0 {
		seasonInfo = p.findallSubTitle(patterns.SeasonPattern, "/[]")
	}

	if len(seasonInfo) > 0 {
		p.seasonTrusted = true
		return p.parseSeason(seasonInfo, true)
	}
	return 1, ""
}

// getUntrustedSeason 获取不可信的季度信息
func (p *TitleMetaParser) getUntrustedSeason() (int, string) {
	// 使用原始的 regexp2 查找（不替换 title）
	// 注意：SEASON_PATTERN_UNTRUSTED 只有一个捕获组，返回字符串列表
	seasonInfoFlat := make([]string, 0)
	text := p.title
	startPos := 0

	for {
		match, err := patterns.SeasonPatternUntrusted.FindStringMatch(text[startPos:])
		if err != nil || match == nil {
			break
		}

		// 提取捕获组
		if len(match.Groups()) > 0 {
			seasonInfoFlat = append(seasonInfoFlat, match.String())
		}

		startPos += match.Index + match.Length
		if startPos >= len(text) {
			break
		}
	}

	// 转换为 [][]string 格式
	if len(seasonInfoFlat) > 0 {
		seasonInfo := make([][]string, len(seasonInfoFlat))
		for i, s := range seasonInfoFlat {
			seasonInfo[i] = []string{s}
		}
		return p.parseSeason(seasonInfo, false)
	}
	return 1, ""
}

// getYear 获取年份信息
func (p *TitleMetaParser) getYear() string {
	yearInfo := p.findallSubTitle(patterns.YearPattern, "[]")
	if len(yearInfo) > 0 && len(yearInfo[0]) > 0 {
		// 去除多余的 () 和 []
		year := strings.ReplaceAll(yearInfo[0][0], "(", "")
		year = strings.ReplaceAll(year, ")", "")
		year = strings.ReplaceAll(year, "[", "")
		year = strings.ReplaceAll(year, "]", "")
		return year
	}
	return ""
}

// nameProcess 处理标题，提取英文、中文和日文名称
func (p *TitleMetaParser) nameProcess() (string, string, string) {
	// TODO: 这里的效果现在并不好, 需要继续优化, 但目前的重点还是在集数解析上
	// 优化 token 处理逻辑
	tempTitle := p.title
	if strings.Contains(tempTitle, "/[]") {
		parts := strings.Split(tempTitle, "/[]")
		// /[] 代表可信的集数或季度, 所以可以相信后面是与集数无用的信息
		// 暂时没有哪个组把集数放前面
		if len(parts) > 1 {
			tempTitle = strings.Join(parts[:len(parts)-1], "[]")
		}
	}
	// 统计 [] 到 max len 后停或到第一个非 []
	tempTitle = strings.TrimSpace(tempTitle)
	count := 0
	maxlen := 0
	flag := false
	for ; maxlen < len(tempTitle); maxlen++ {
		if tempTitle[maxlen] == '[' || tempTitle[maxlen] == ']' {
			count++
			if count >= 10 && flag {
				break
			}
		} else {
			flag = true
		}
	}

	// 用 [\[\]] 分割 - 按 [ 或 ] 分割
	p.token = strings.FieldsFunc(tempTitle[:maxlen], func(r rune) bool {
		return r == '[' || r == ']'
	})

	// 过滤掉空白字符的 token
	tokenFiltered := make([]string, 0)
	for _, token := range p.token {
		if strings.TrimSpace(token) != "" {
			tokenFiltered = append(tokenFiltered, token)
		}
	}
	p.token = tokenFiltered

	if len(p.token) > 5 {
		p.token = p.token[:5]
	}

	tokenPriority := make([]int, len(p.token))
	for i, s := range p.token {
		tokenPriority[i] = len(s)
	}

	var animeTitle string

	if len(p.token) == 1 {
		animeTitle = p.token[0]
	} else if len(p.token) == 2 {
		animeTitle = p.token[1]
	} else if len(p.token) > 2 {
		tokenPriority[1] += 4
		for idx := 0; idx < 3 && idx < len(p.token); idx++ {
			token := p.token[idx]
			if strings.Contains(token, "/") {
				tokenPriority[idx] += 10
			}
			if strings.Contains(token, "&") {
				tokenPriority[idx] -= 12
			}
			if strings.Contains(token, "字幕") {
				tokenPriority[idx] -= 90
			}

			// 英文匹配
			if hasEnglish(token) {
				tokenPriority[idx] += 2
			}

			// 日文匹配
			jpCount := countJapanese(token)
			if jpCount >= 2 {
				tokenPriority[idx] += jpCount * 2
			}

			// 中文匹配
			cnCount := countChinese(token)
			if cnCount >= 2 {
				tokenPriority[idx] += cnCount * 2
			}
		}

		idx := 0
		maxPriority := tokenPriority[0]
		for i := 1; i < len(tokenPriority); i++ {
			if tokenPriority[i] > maxPriority {
				maxPriority = tokenPriority[i]
				idx = i
			}
		}
		animeTitle = p.token[idx]
	}

	animeTitle = strings.Trim(animeTitle, "\\")
	animeTitle = strings.TrimSpace(animeTitle)

	nameEn, nameZh, nameJp := "", "", ""

	// 分割标题 - 把多种分隔符统一替换成 /，然后按 / 分割
	temp := animeTitle
	temp = strings.ReplaceAll(temp, "  ", "/")  // 两个空格
	temp = strings.ReplaceAll(temp, "-  ", "/") // 破折号+两个空格
	split := strings.Split(temp, "/")

	// 移除空字符串
	filtered := make([]string, 0)
	for _, s := range split {
		if s != "" {
			filtered = append(filtered, s)
		}
	}
	split = filtered

	if len(split) == 1 {
		// 主要的思想就是从头或者尾部找出一个中文名
		cnCount := countChinese(split[0])
		chineseRatio := 0.0
		runeCount := len([]rune(split[0]))
		if runeCount > 0 {
			chineseRatio = float64(cnCount) / float64(runeCount)
		}

		if chineseRatio <= 0.7 {
			splitSpace := strings.Split(split[0], " ")

			for _, idx := range []int{0, len(splitSpace) - 1} {
				if idx >= 0 && idx < len(splitSpace) {
					if startsWithChinese(splitSpace[idx]) {
						chs := splitSpace[idx]
						newSplit := make([]string, 0)
						for _, s := range splitSpace {
							if s != chs {
								newSplit = append(newSplit, s)
							}
						}
						split = []string{chs, strings.Join(newSplit, " ")}
						break
					}
				}
			}
		}
	}

	for _, token := range split {
		if hasJapanese(token) && nameJp == "" {
			nameJp = strings.TrimSpace(token)
		} else if hasChinese(token) && nameZh == "" {
			nameZh = strings.TrimSpace(token)
		} else if hasEnglish(token) && nameEn == "" {
			nameEn = strings.TrimSpace(token)
		}
	}

	return nameEn, nameZh, nameJp
}

// getGroup 获取字幕组信息
func (p *TitleMetaParser) getGroup() string {
	for _, group := range p.token {
		trimmed := strings.TrimSpace(group)
		if trimmed != "" {
			trimmed = strings.ReplaceAll(trimmed, "/", "")
			return strings.TrimSpace(trimmed)
		}
	}
	return ""
}

// getVideoInfo 获取视频格式信息
func (p *TitleMetaParser) getVideoInfo() []string {
	matches := p.findallSubTitle(patterns.VideoTypePattern, "[]")
	result := make([]string, 0)
	for _, match := range matches {
		if len(match) > 0 && match[0] != "" {
			result = append(result, match[0])
		}
	}
	return result
}

// getResolutionInfo 获取分辨率信息
func (p *TitleMetaParser) getResolutionInfo() []string {
	matches := p.findallSubTitle(patterns.ResolutionPatternTrust, "[]")
	result := make([]string, 0)
	for _, match := range matches {
		if len(match) > 0 && match[0] != "" {
			result = append(result, match[0])
		}
	}
	return result
}

// getSourceInfo 获取视频来源信息
func (p *TitleMetaParser) getSourceInfo() []string {
	matches := p.findallSubTitle(patterns.SourceRe, "[]")
	result := make([]string, 0)
	for _, match := range matches {
		if len(match) > 0 && match[0] != "" {
			result = append(result, match[0])
		}
	}
	return result
}

// getUnusefulInfo 获取无用信息
func (p *TitleMetaParser) getUnusefulInfo() []string {
	matches := p.findallSubTitle(patterns.UnusefulRe, "[]")
	result := make([]string, 0)
	for _, match := range matches {
		if len(match) > 0 && match[0] != "" {
			result = append(result, match[0])
		}
	}
	return result
}

// getSubtitleType 获取字幕类型
func (p *TitleMetaParser) getSubtitleType() string {
	matches := p.findallSubTitle(patterns.SubReType, "[]")
	s := ""
	for _, match := range matches {
		if len(match) > 0 && match[0] != "" {
			sub := match[0]
			if !strings.Contains(s, sub) {
				s += sub
			}
		}
	}
	return s
}

// getSubtitleLanguage 获取字幕信息
func (p *TitleMetaParser) getSubtitleLanguage() string {
	sub := ""

	if len(p.findallSubTitle(patterns.SubReChs, "[]")) > 0 {
		sub += "简"
	}

	if len(p.findallSubTitle(patterns.SubReCht, "[]")) > 0 {
		sub += "繁"
	}

	if len(p.findallSubTitle(patterns.SubReJp, "[]")) > 0 {
		sub += "日"
	}

	if len(p.findallSubTitle(patterns.SubReEnglish, "[]")) > 0 {
		sub += "英"
	}

	return sub
}

// getAudioInfo 获取音频信息
func (p *TitleMetaParser) getAudioInfo() string {
	matches := p.findallSubTitle(patterns.AudioInfo, "[]")
	if len(matches) > 0 && len(matches[0]) > 0 {
		return matches[0][0]
	}
	return ""
}

// Parse 解析标题，返回 Episode 信息
func (p *TitleMetaParser) Parse(title string) *model.EpisodeMetadata {
	meta := p.ParseEpisode(title)
	// 下面的解析名字没有要放出去
	group := meta.Group

	nameEn, nameZh, nameJp := p.nameProcess()
	titleRaw := firstNonEmptyString(nameZh, nameJp, nameEn)

	if group == "" {
		group = p.getGroup()
		// 当 group 被包含在 title 中, 则清空 group
		if group != "" && (strings.Contains(nameEn, group) || strings.Contains(nameZh, group) || strings.Contains(nameJp, group)) {
			group = ""
		}
	}

	meta.Title = titleRaw
	meta.Group = group

	return meta
}

func (p *TitleMetaParser) getVersion() int {
	versionInfo := p.findallSubTitle(patterns.VersionPattern, "[]")
	if len(versionInfo) == 0 {
		versionInfo = p.findallSubTitle(patterns.VersionWithNum, "[]")
	}
	if len(versionInfo) > 0 {
		return p.episodeInfoToEpisode(versionInfo[0])
	}
	return 1
}

// ParseEpisode 解析视频的集数
func (p *TitleMetaParser) ParseEpisode(title string) *model.EpisodeMetadata {
	ep := &model.EpisodeMetadata{}
	p.rawTitle = title
	p.title = title
	p.title = utils.ProcessTitle(p.title)

	// 末尾加一个 / 处理边界
	p.title += "/"
	// 开头加一个[ 处理边界
	p.title = "[" + p.title
	// 从一个自己定义的字幕组文件中获取字幕组信息, 保证字幕组信息的准确性
	// TODO: 这个后面要放成一个可更新的文件, 现在先这样写着
	ep.Group = p.getGroupInfo()
	ep.Year = p.getYear()
	sourceInfo := p.getSourceInfo()
	resolutionInfo := p.getResolutionInfo()
	ep.AudioInfo = p.getAudioInfo()
	videoInfo := p.getVideoInfo()

	// 要先拿字幕类型, 双语什么的会影响字幕语言的判断
	ep.SubType = p.getSubtitleType()
	ep.Sub = p.getSubtitleLanguage()
	// 无用信息后面也要做成一个可更新的文件, 着实情况太多了
	_ = p.getUnusefulInfo() // 清理无用信息，但不使用结果

	// 先排除 range 的集数, 再排除可信的集数, 最后才是非可信的集数
	// 用episode = -1 来表示全集
	ep.Collection, ep.EpisodeStart, ep.EpisodeEnd = p.getCollectionInfo()
	ep.Version = p.getVersion()

	// 处理可信的集数和季度, collection 的季度和集数解析没有意义
	if ep.Collection { // 是合集，episode = -1
		ep.Episode = -1
	} else {
		// 不是合集，尝试获取可信集数
		ep.Episode = p.getTrustedEpisode()
		if ep.Episode == -1 {
			// 没有可信集数，获取不可信集数
			ep.Episode = p.getUntrustedEpisode()
		}
	}

	// 开始解析 季度的信息
	season, seasonRaw := p.getTrustedSeason()

	if !p.seasonTrusted {
		season, seasonRaw = p.getUntrustedSeason()
	}
	ep.Season = season
	ep.SeasonRaw = seasonRaw

	if len(sourceInfo) > 0 {
		ep.Source = sourceInfo[0]
	}

	if len(resolutionInfo) > 0 {
		ep.Resolution = resolutionInfo[0]
	}

	if len(videoInfo) > 0 {
		ep.VideoInfo = strings.Join(videoInfo, ",")
	}

	return ep
}

// IsV1 判断是否是 v1 番剧
func IsV1(title string) bool {
	match, _ := patterns.VersionPattern.FindStringMatch(title)
	return match != nil
}

// IsPoint5 判断是否是 .5 番剧
func IsPoint5(title string) bool {
	match, _ := patterns.Point5Re.FindStringMatch(title)
	return match != nil
}

// ============ 字符判断辅助函数（替代正则提升性能）============

// isChinese 判断字符是否为中文
func isChinese(r rune) bool {
	return r >= 0x4E00 && r <= 0x9FFF
}

// isJapanese 判断字符是否为日文
func isJapanese(r rune) bool {
	return (r >= 0x3040 && r <= 0x309F) || (r >= 0x30A0 && r <= 0x30FF)
}

// isEnglish 判断字符是否为英文字母
func isEnglish(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

// countChinese 统计中文字符数量
func countChinese(s string) int {
	count := 0
	for _, r := range s {
		if isChinese(r) {
			count++
		}
	}
	return count
}

// countJapanese 统计日文字符数量
func countJapanese(s string) int {
	count := 0
	for _, r := range s {
		if isJapanese(r) {
			count++
		}
	}
	return count
}

// hasChinese 判断是否包含至少2个中文字符
func hasChinese(s string) bool {
	count := 0
	for _, r := range s {
		if isChinese(r) {
			count++
			if count >= 2 {
				return true
			}
		}
	}
	return false
}

// hasJapanese 判断是否包含至少2个日文字符
func hasJapanese(s string) bool {
	count := 0
	for _, r := range s {
		if isJapanese(r) {
			count++
			if count >= 2 {
				return true
			}
		}
	}
	return false
}

// hasEnglish 判断是否包含至少3个连续英文字母
func hasEnglish(s string) bool {
	count := 0
	for _, r := range s {
		if isEnglish(r) {
			count++
			if count >= 3 {
				return true
			}
		} else {
			count = 0
		}
	}
	return false
}

// startsWithChinese 判断是否以至少2个中文字符开头
func startsWithChinese(s string) bool {
	count := 0
	for _, r := range s {
		if isChinese(r) {
			count++
			if count >= 2 {
				return true
			}
		} else {
			return false
		}
	}
	return count >= 2
}

func firstNonEmptyString(strs ...string) string {
	for _, s := range strs {
		if s != "" {
			return s
		}
	}
	return ""
}
