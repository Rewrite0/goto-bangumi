package patterns

import "github.com/dlclark/regexp2"

// 字幕边界定义
const (
	SubBoundaryStart = `[★／/_&（）\s\-\.\[\]\(\)简繁中日英体字體]`
	SubBoundaryEnd   = `(?=[★／/_&（）\s\-\.\[\]\(\)简繁中日英体字體])`
)

// SubReCht 繁体中文字幕匹配
var SubReCht = regexp2.MustCompile(
	SubBoundaryStart+`
    (CHT
    |繁
    |BIG5
    )
    `+SubBoundaryEnd,
	regexp2.IgnoreCase|regexp2.IgnorePatternWhitespace,
)

// SubReChs 简体中文字幕匹配
var SubReChs = regexp2.MustCompile(
	SubBoundaryStart+`
    (CHS
    |SC
    |简
    |GB
    |GBJP)
    `+SubBoundaryEnd,
	regexp2.IgnoreCase|regexp2.IgnorePatternWhitespace,
)

// SubReJp 日文字幕匹配
var SubReJp = regexp2.MustCompile(
	SubBoundaryStart+`
    (JP
    |GBJP
    |JPN
    |日)
    `+SubBoundaryEnd,
	regexp2.IgnoreCase|regexp2.IgnorePatternWhitespace,
)

// SubReEnglish 英文字幕匹配
var SubReEnglish = regexp2.MustCompile(
	SubBoundaryStart+`
    ( 英)
    `+SubBoundaryEnd,
	regexp2.IgnoreCase|regexp2.IgnorePatternWhitespace,
)

// SubReType 字幕类型匹配
var SubReType = regexp2.MustCompile(
	`
    (外挂
    |内封
    |内嵌
    |硬字幕
    |软字幕
    |ASS
    |SRT
    |双语)
    `,
	regexp2.IgnoreCase|regexp2.IgnorePatternWhitespace,
)
