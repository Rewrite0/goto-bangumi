package model

import "fmt"

// Episode 表示解析后的剧集信息
type Episode struct {
	TitleRaw   string
	Season     int
	SeasonRaw  string
	Episode    int
	Sub        string
	SubType    string
	Group      string
	Year       string
	Resolution string
	Source     string
	AudioInfo  string
	VideoInfo  string
}

// String 返回 Episode 的格式化字符串
func (e *Episode) String() string {
	return fmt.Sprintf(`Episode 解析结果:
  标题: %s
  季度: S%02d (%s)
  集数: E%02d
  字幕组: %s
  字幕语言: %s
  字幕类型: %s
  分辨率: %s
  来源: %s
  年份: %s
  音频: %s
  视频: %s`,
		e.TitleRaw,
		e.Season, e.SeasonRaw, e.Episode,
		e.Group, e.Sub, e.SubType,
		e.Resolution, e.Source, e.Year,
		e.AudioInfo, e.VideoInfo)
}
