package patterns

import "github.com/dlclark/regexp2"

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
	BoundaryStart+`# Frame rate
    (AAC(?:x2)?
    |AAC(?:2\.0)?
    |FLAC(?:x2)?
    |DDP(?:2\.0)?
    |OPUS
    )
    `+BoundaryEnd,
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
