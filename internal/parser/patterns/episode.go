package patterns

import "github.com/dlclark/regexp2"

// EpisodePatternTrustWithBoundary 可信集数匹配（带边界）
var EpisodePatternTrustWithBoundary = regexp2.MustCompile(
	BoundaryStart+`
    (
    E(\d+?) # E9
    |(\d+?).?END # 9END  9 END
    |(\d+?)pre) # 9pre
    `+BoundaryEnd,
	regexp2.IgnoreCase|regexp2.IgnorePatternWhitespace,
)

// EpisodePatternTrust 可信集数匹配（无边界）
var EpisodePatternTrust = regexp2.MustCompile(
	`
	(?:第?(\d+?|[一二三四五六七八九十]+)[话話集]  #第12话 第12集
    |EP(\d+) # EP12
		|\[(\d+?)\] # [12]
    |S\d+(?:EP?(\d+)) # S1EP12 S01E12
    |(\d+)v\d # 12v2
    |(\d+?)Fin # 12Fin
    |-\s(\d+)\s # - 12
    )
`,
	regexp2.IgnoreCase|regexp2.IgnorePatternWhitespace,
)

// ============ 合集匹配规则（拆分） ============

// CollectionRangeRe 数字范围匹配 如 X01-12（X为非S字符）
var CollectionRangeRe = regexp2.MustCompile(
	`[^sS]([\d]{2,})\s?[-~]\s?([\d]{2,})`,
	regexp2.IgnoreCase,
)

// CollectionBracketRe 方括号范围匹配 如 [01-12]
var CollectionBracketRe = regexp2.MustCompile(
	`\[(\d+)-(\d+)\]`,
	regexp2.IgnoreCase,
)

// CollectionZhRe 中文范围匹配 如 第01-12话
var CollectionZhRe = regexp2.MustCompile(
	`[第]([\d]{2,})\s?[-~]\s?([\d]{2,})\s?[話话集]`,
	regexp2.IgnoreCase,
)

// CollectionZh2Re 中文范围匹配2 如 01-12话
var CollectionZh2Re = regexp2.MustCompile(
	`第?(\d+)\s?[-~]\s?(\d+)[話话集]`,
	regexp2.IgnoreCase,
)

// CollectionAllZhRe 全集匹配 如 全12话
var CollectionAllZhRe = regexp2.MustCompile(
	`[全](\d+)?[話话集]`,
	regexp2.IgnoreCase,
)

// CollectionVolRe vol匹配 如 vol.1
var CollectionVolRe = regexp2.MustCompile(
	`vol\.(\d+)`,
	regexp2.IgnoreCase,
)

// CollectionVolRangeRe vol范围匹配 如 vol.1-2
var CollectionVolRangeRe = regexp2.MustCompile(
	`vol\.(\d+)[-~_](\d+)`,
	regexp2.IgnoreCase,
)

// CollectionSeasonRangeRe 季度范围匹配 如 S1-S2
var CollectionSeasonRangeRe = regexp2.MustCompile(
	`S(\d+)[-~+]S(\d+)`,
	regexp2.IgnoreCase,
)

// CollectionRangePatterns 有范围的合集规则（返回 start, end）
var CollectionRangePatterns = []*regexp2.Regexp{
	CollectionZhRe,        // 第01-12话（优先级最高）
	CollectionZh2Re,       // 01-12话
	CollectionBracketRe,   // [01-12]
	CollectionVolRangeRe,  // vol.1-2
	CollectionSeasonRangeRe, // S1-S2
	CollectionRangeRe,     // X01-12（优先级最低，容易误匹配）
}

// CollectionSinglePatterns 无范围的合集规则（只标识是合集）
var CollectionSinglePatterns = []*regexp2.Regexp{
	CollectionAllZhRe, // 全12话
	CollectionVolRe,   // vol.1
}

// EpisodeReUntrusted 不可信集数匹配
var EpisodeReUntrusted = regexp2.MustCompile(
	BoundaryStart+`
        ((\d+?))
        `+BoundaryEnd,
	regexp2.IgnorePatternWhitespace,
)

// VersionPattern V1 版本匹配
var VersionPattern = regexp2.MustCompile(
	BoundaryStart+`
	(?:v\d+?)
    `+BoundaryEnd,
	regexp2.IgnoreCase|regexp2.IgnorePatternWhitespace,
)

var VersionWithNum = regexp2.MustCompile(
	`(?<=\d)(v(\d+?))`,  // 用 lookbehind，只匹配 v2 部分，保留前面的集数
	regexp2.IgnoreCase,
)
