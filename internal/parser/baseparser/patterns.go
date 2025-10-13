package baseparser

import (
	"regexp"

	"github.com/dlclark/regexp2"
)

// 边界字符定义
const (
	splitPattern  = `★／/_&（）\s\-\.\[\]\(\)`
	boundaryStart = `[` + splitPattern + `]`
	boundaryEnd   = `(?=[` + splitPattern + `])`
)

// ChineseNumberMap 中文数字到阿拉伯数字的映射
var ChineseNumberMap = map[string]int{
	"一": 1, "二": 2, "三": 3, "四": 4, "五": 5,
	"六": 6, "七": 7, "八": 8, "九": 9, "十": 10,
}

var ChineseNumberUpperMap = map[string]int{
	"零": 0, "壹": 1, "贰": 2, "叁": 3, "肆": 4,
	"伍": 5, "陆": 6, "柒": 7, "捌": 8, "玖": 9,
}

// RomanNumbers 罗马数字到阿拉伯数字的映射
var RomanNumbers = map[string]int{
	"I": 1, "II": 2, "III": 3, "IV": 4, "V": 5,
}

// ============ Episode 相关正则 ============

// EpisodePatternTruestWithBoundary 可信集数匹配（带边界）
var EpisodePatternTruestWithBoundary = regexp2.MustCompile(
	boundaryStart+`
    (
    E(\d+?)
    |(\d+?).?END
    |(\d+?)pre)
    `+boundaryEnd,
	regexp2.IgnoreCase|regexp2.IgnorePatternWhitespace,
)

// EpisodePatternTruest 可信集数匹配（无边界）
var EpisodePatternTruest = regexp2.MustCompile(
	`
	(?:第?(\d+)[话話集]
    |EP(\d+)
    |S\d+(?:EP?(\d+))
    |-\s(\d+)
    |(\d+)v\d
    |(\d+?)Fin # -12Fin
    )
`,
	regexp2.IgnoreCase|regexp2.IgnorePatternWhitespace,
)

// CollectionPattern 合集匹配
var CollectionPattern = regexp2.MustCompile(
	`
		[^sS]([\d]{2,})\s?[-~]\s?([\d]{2,})
		|\[(\d+)-(\d+)]
		|[第]([\d]{2,})\s?[-~]\s?([\d]{2,})\s?[話话集]
		|[全]([\d-]*?)[話话集]
		|第?(\d*)\s?[-~]\s(\d*)[話话集]
		|vol\.\d
		|vol\.\d[-~_]\d
		|S\d[-~+]S\d #  S1-+S2
	`,
	regexp2.IgnoreCase|regexp2.IgnorePatternWhitespace,
)

// EpisodeReUntrusted 不可信集数匹配
var EpisodeReUntrusted = regexp2.MustCompile(
	boundaryStart+`
        ((\d+?))
        `+boundaryEnd,
	regexp2.IgnorePatternWhitespace,
)

// ============ Season 相关正则 ============

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
	boundaryStart+`
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

var MikanSeasonPaattern = regexp.MustCompile(`\s(?:第(.)季|(贰))$`)

// ============ 字幕组相关 ============

// GroupRe 字幕组名称匹配
var GroupRe = regexp2.MustCompile(
	boundaryStart+`
    (ANi
    |LoliHouse
    |SweetSub
    |Pre-S
    |H-Enc
    |TOC
    |Billion Meta Lab
    |Lilith-Raws
    |DBD-Raws
    |NEO·QSW
    |SBSUB
    |MagicStar
    |7³ACG
    |KitaujiSub
    |Doomdos
    |Prejudice-Studio
    |GM-Team
    |VCB-Studio
    |神椿观测站
    |极影字幕社
    |百冬练习组
    |猎户手抄部
    |喵萌奶茶屋
    |萌樱字幕组
    |三明治摆烂组
    |绿茶字幕组
    |梦蓝字幕组
    |幻樱字幕组
    |织梦字幕组
    |北宇治字组
    |北宇治字幕组
    |霜庭云花Sub
    |氢气烤肉架
    |豌豆字幕组
    |豌豆
    |DBD
    |风之圣殿字幕组
    |黒ネズミたち
    |桜都字幕组
    |漫猫字幕组
    |猫恋汉化组
    |黑白字幕组
    |猎户压制部
    |猎户手抄部
    |沸班亚马制作组
    |星空字幕组
    |光雨字幕组
    |樱桃花字幕组
    |动漫国字幕组
    |动漫国
    |千夏字幕组
    |SW字幕组
    |澄空学园
    |华盟字幕社
    |诸神字幕组
    |雪飘工作室
    |❀拨雪寻春❀
    |夜莺家族
    |YYQ字幕组
    |APTX4869
    |Prejudice-Studio
    |丸子家族
    )
    `+boundaryEnd,
	regexp2.IgnorePatternWhitespace,
)

// ============ 视频/音频/分辨率相关 ============

// VideoTypePattern 视频编码格式匹配
var VideoTypePattern = regexp2.MustCompile(
	`
    # Frame rate
    ( # Video codec
    8-?BITS?
    |10-?BITS?
    |HI10P?
    |[HX].?26[4|5]
    |AVC
    |HEVC2?
    # Video format
    |AVI
    |AV1
    |RMVB
    |MKV
    |MP4
    # video quailty
    |HD
    |UHD
    |SRT[x2].?
    |ASS[x2].? # AAAx2
    |PGS
    |V[123]
    |Remux
    |OVA)
    `,
	regexp2.IgnoreCase|regexp2.IgnorePatternWhitespace,
)

// FrameRate 帧率匹配
var FrameRate = regexp2.MustCompile(
	`
    (23.976FPS
    |24FPS
    |29.97FPS
    |[30|60|120]FPS
    )
    `,
	regexp2.IgnoreCase|regexp2.IgnorePatternWhitespace,
)

// DecodeInfo 解码信息匹配
var DecodeInfo = regexp2.MustCompile(
	`
    (HEVC(?:-10bit)?
    |AVC
    |H[\.|x]?264
    |H[\.|x]?265
    |X264
    |X265
    |AV1
    )
    `,
	regexp2.IgnoreCase|regexp2.IgnorePatternWhitespace,
)

// AudioInfo 音频编码匹配
var AudioInfo = regexp2.MustCompile(
	boundaryStart+`# Frame rate
    (AAC(?:x2)?
    |AAC(?:2\.0)?
    |FLAC(?:x2)?
    |DDP(?:2\.0)?
    |OPUS
    )
    `+boundaryEnd,
	regexp2.IgnoreCase|regexp2.IgnorePatternWhitespace,
)

// ResolutionPatternTrust 可信分辨率匹配
var ResolutionPatternTrust = regexp2.MustCompile(
	`
    (\d{3,4}[×xX]\d{3,4}
    |1080p
    |720p
    |480p
    |2160p
    |4K
    )
    `,
	regexp2.IgnoreCase|regexp2.IgnorePatternWhitespace,
)

// ============ 来源和年份 ============

// SourceRe 视频来源匹配
var SourceRe = regexp2.MustCompile(
	boundaryStart+`
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
    `+boundaryEnd,
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

// ============ 字幕相关 ============

const (
	subBoundaryStart = `[★／/_&（）\s\-\.\[\]\(\)简繁中日英体字體]`
	subBoundaryEnd   = `(?=[★／/_&（）\s\-\.\[\]\(\)简繁中日英体字體])`
)

// SubReCht 繁体中文字幕匹配
var SubReCht = regexp2.MustCompile(
	subBoundaryStart+`
    (CHT
    |繁
    |BIG5
    )
    `+subBoundaryEnd,
	regexp2.IgnoreCase|regexp2.IgnorePatternWhitespace,
)

// SubReChs 简体中文字幕匹配
var SubReChs = regexp2.MustCompile(
	subBoundaryStart+`
    (CHS
    |SC
    |简
    |GB
    |GBJP)
    `+subBoundaryEnd,
	regexp2.IgnoreCase|regexp2.IgnorePatternWhitespace,
)

// SubReJp 日文字幕匹配
var SubReJp = regexp2.MustCompile(
	subBoundaryStart+`
    (JP
    |GBJP
    |日)
    `+subBoundaryEnd,
	regexp2.IgnoreCase|regexp2.IgnorePatternWhitespace,
)

// SubReEnglish 英文字幕匹配
var SubReEnglish = regexp2.MustCompile(
	subBoundaryStart+`
    ( 英)
    `+subBoundaryEnd,
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

// ============ 其他辅助正则 ============

// UnusefulRe 无用信息匹配
var UnusefulRe = regexp2.MustCompile(
	`(?<=`+boundaryStart+`)
        ( .?[\d一四七十春夏秋冬季]{1,2}月(新番|短剧).*?
        | 港澳台地区
        | 国漫
        | END
        | 招募.*?
        | \d{4}年\d{1,2}月.*? # 2024年1月
        | \d{4}\.\d{1,2}\.\d{1,2}
        | Vol\.\d-\d #1-6
        |[网盘无水印高清下载迅雷]{4,10})
        `+boundaryEnd,
	regexp2.IgnorePatternWhitespace,
)

// V1Re V1 版本匹配
var V1Re = regexp2.MustCompile(
	boundaryStart+`
    (V1)
    `+boundaryEnd,
	regexp2.IgnoreCase|regexp2.IgnorePatternWhitespace,
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
