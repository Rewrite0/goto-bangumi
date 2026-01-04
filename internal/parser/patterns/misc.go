package patterns

import "github.com/dlclark/regexp2"

// UnusefulRe 无用信息匹配
var UnusefulRe = regexp2.MustCompile(
	`(?<=`+BoundaryStart+`)
        ( .?[\d一四七十春夏秋冬季]{1,2}月(新番|短剧).*?
        | 港澳台地区
        | 国漫
        | END
        | 招募.*?
        | \d{4}年\d{1,2}月.*? # 2024年1月
        | \d{4}\.\d{1,2}\.\d{1,2}
        | Vol\.\d-\d #1-6
        |[网盘无水印高清下载迅雷]{4,10})
        `+BoundaryEnd,
	regexp2.IgnorePatternWhitespace,
)


// Point5Re 半集（如 12.5 集）匹配
var Point5Re = regexp2.MustCompile(
	`(第?\d+?\.\d+?[话話集]
    |EP?\d+?\.\d+?
    |-\s\d+?\.\d+?
    |\d+?\.\d+?v\d+?
    |\d+?\.\d+?(END|pre)
    )
    (?=[\s_\-\[\]$\.\(\)])
`,
	regexp2.IgnoreCase|regexp2.IgnorePatternWhitespace,
)
