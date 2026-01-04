package patterns

import (
	"regexp"

	"github.com/dlclark/regexp2"
)

// SeasonPatternTruest 最可信季度匹配
var SeasonPatternTruest = regexp2.MustCompile(
	`
    (第(.{1,3})季       # 匹配"第...季"格式
    |第(.{1,3})期        # 匹配"第...期"格式
    |第.{1,3}部分      # 匹配"第...部分"格式
    |[Ss]eason\s?(\d{1,2})  # 匹配"Season X"格式
    |SEASON\s?(\d{1,2})  # 匹配"SEASON X"格式
    )
    `,
	regexp2.IgnorePatternWhitespace,
)

// SeasonPattern 可信季度匹配（带边界）
var SeasonPattern = regexp2.MustCompile(
	BoundaryStart+`
    ([Ss](\d{1,2})         # 匹配"SX"格式
    |(\d+)[r|n]d(?:\sSeason)?  # 匹配"Xnd Season"格式
    |part \d   #part 6
    |(IV|III|II|I)            # 匹配罗马数字
    ) (?=[\s_\.\-\[\]/\)\($E])  # 结束边界（不消耗）
    `,
	regexp2.IgnorePatternWhitespace,
)

// SeasonPatternUntrusted 不可信季度匹配
var SeasonPatternUntrusted = regexp2.MustCompile(`\d+(?!\.)`, regexp2.None)

// MikanSeasonPattern Mikan 季度匹配
var MikanSeasonPattern = regexp.MustCompile(`\s(?:第(.)季|(贰))$`)
