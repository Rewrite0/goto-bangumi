package patterns

import "github.com/dlclark/regexp2"

// SourceRe 视频来源匹配
var SourceRe = regexp2.MustCompile(
	BoundaryStart+`
    (B-Global
    |Baha
    |Bilibili
    |AT-X
    |W[eE][Bb]-?(?:Rip)?(?:DL)? # WEBRIP 和 WEBDL
    |CR
    |ABEMA
    |BD(?:RIP)?
    |JPBD
    |viutv[粤语]*?)
    `+BoundaryEnd,
	regexp2.IgnoreCase|regexp2.IgnorePatternWhitespace,
)

// YearPattern 年份匹配
var YearPattern = regexp2.MustCompile(
	`
    (
    \(19\d{2}\)
    |\(20\d{2}\) # (1900) (2000)
    |\[20\d{2}\] # [2020] 对 GM-Team 特化
    )
    `,
	regexp2.IgnorePatternWhitespace,
)
