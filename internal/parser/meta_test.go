package parser

import (
	"testing"
)

func TestRawParser(t *testing.T) {
	tests := []struct {
		name             string
		content          string
		wantGroup        string
		wantTitleRaw     string
		wantRes          string
		wantEp           int
		wantSeason       int
		wantSub          string
		wantCollection   bool
		wantVersion      int
		wantEpisodeStart int
		wantEpisodeEnd   int
	}{
		{
			name:         "幻樱字幕组 - 古见同学",
			content:      "【幻樱字幕组】【4月新番】【古见同学有交流障碍症 第二季 Komi-san wa, Komyushou Desu. S02】【22】【GB_MP4】【1920X1080】",
			wantGroup:    "幻樱字幕组",
			wantTitleRaw: "古见同学有交流障碍症",
			wantRes:      "1920X1080",
			wantEp:       22,
			wantSeason:   2,
			wantSub:      "简",
		},
		{
			name:         "百冬练习组&LoliHouse - BanG Dream",
			content:      "[百冬练习组&LoliHouse] BanG Dream! 少女乐团派对！☆PICO FEVER！ / Garupa Pico: Fever! - 26 [WebRip 1080p HEVC-10bit AAC][简繁内封字幕][END] [101.69 MB]",
			wantGroup:    "百冬练习组&LoliHouse",
			wantTitleRaw: "BanG Dream! 少女乐团派对！☆PICO FEVER！",
			wantRes:      "1080p",
			wantEp:       26,
			wantSeason:   1,
			wantSub:      "简繁",
		},
		{
			name:         "喵萌奶茶屋 - 夏日重现",
			content:      "【喵萌奶茶屋】★04月新番★[夏日重现/Summer Time Rendering][11][1080p][繁日双语][招募翻译]",
			wantGroup:    "喵萌奶茶屋",
			wantTitleRaw: "夏日重现",
			wantRes:      "1080p",
			wantEp:       11,
			wantSeason:   1,
			wantSub:      "繁日",
		},
		{
			name:         "Lilith-Raws - 天使",
			content:      "[Lilith-Raws] 关于我在无意间被隔壁的天使变成废柴这件事 / Otonari no Tenshi-sama - 09 [Baha][WEB-DL][1080p][AVC AAC][CHT][MP4]",
			wantGroup:    "Lilith-Raws",
			wantTitleRaw: "关于我在无意间被隔壁的天使变成废柴这件事",
			wantRes:      "1080p",
			wantEp:       9,
			wantSeason:   1,
			wantSub:      "繁",
		},
		{
			name:         "梦蓝字幕组 - 哆啦A梦",
			content:      "[梦蓝字幕组]New Doraemon 哆啦A梦新番[747][2023.02.25][AVC][1080P][GB_JP][MP4]",
			wantGroup:    "梦蓝字幕组",
			wantTitleRaw: "哆啦A梦新番",
			wantRes:      "1080P",
			wantEp:       747,
			wantSeason:   1,
			wantSub:      "简日",
		},
		{
			name:         "织梦字幕组 - 尼尔",
			content:      "[织梦字幕组][尼尔：机械纪元 NieR Automata Ver1.1a][02集][1080P][AVC][简日双语]",
			wantGroup:    "织梦字幕组",
			wantTitleRaw: "尼尔：机械纪元",
			wantRes:      "1080P",
			wantEp:       2,
			wantSeason:   1,
			wantSub:      "简日",
		},
		{
			name:         "MagicStar - 假面骑士",
			content:      "[MagicStar] 假面骑士Geats / 仮面ライダーギーツ EP33 [WEBDL] [1080p] [TTFC]【生】",
			wantGroup:    "MagicStar",
			wantTitleRaw: "假面骑士Geats",
			wantRes:      "1080p",
			wantEp:       33,
			wantSeason:   1,
		},
		{
			name:         "极影字幕社 - 天国大魔境",
			content:      "【极影字幕社】★4月新番 天国大魔境 Tengoku Daimakyou 第05话 GB 720P MP4（字幕社招人内详）",
			wantTitleRaw: "天国大魔境",
			wantRes:      "720P",
			wantEp:       5,
			wantSeason:   1,
			wantGroup:    "极影字幕社",
			wantSub:      "简",
		},
		{
			name:         "喵萌奶茶屋 - 银砂糖师",
			content:      "【喵萌奶茶屋】★07月新番★[银砂糖师与黑妖精 ~ Sugar Apple Fairy Tale ~][13][1080p][简日双语][招募翻译]",
			wantGroup:    "喵萌奶茶屋",
			wantTitleRaw: "银砂糖师与黑妖精",
			wantRes:      "1080p",
			wantEp:       13,
			wantSeason:   1,
			wantSub:      "简日",
		},
		{
			name:         "ANi - 16bit",
			content:      "[ANi]  16bit 的感动 ANOTHER LAYER - 02 [1080P][Baha][WEB-DL][AAC AVC][CHT][MP4]",
			wantGroup:    "ANi",
			wantTitleRaw: "16bit 的感动 ANOTHER LAYER",
			wantRes:      "1080P",
			wantEp:       2,
			wantSeason:   1,
			wantSub:      "繁",
		},
		{
			name:         "Billion Meta Lab - 终末列车",
			content:      "[Billion Meta Lab] 终末列车寻往何方 Shuumatsu Torein Dokoe Iku [12][1080][HEVC 10bit][简繁日内封][END]",
			wantGroup:    "Billion Meta Lab",
			wantTitleRaw: "终末列车寻往何方",
			wantEp:       12,
			wantSeason:   1,
		},
		{
			name:         "超超超超超喜欢你 - 第二季",
			content:      "【1月】超超超超超喜欢你的100个女朋友 第二季 07.mp4",
			wantGroup:    "1月",
			wantTitleRaw: "超超超超超喜欢你的100个女朋友",
			wantEp:       7,
			wantSeason:   2,
		},
		{
			name:         "LoliHouse - 2.5次元",
			content:      "[LoliHouse] 2.5次元的诱惑 / 2.5-jigen no Ririsa - 01 [WebRip 1080p HEVC-10bit AAC][简繁内封字幕][609.59 MB]",
			wantGroup:    "LoliHouse",
			wantTitleRaw: "2.5次元的诱惑",
			wantRes:      "1080p",
			wantEp:       1,
			wantSub:      "简繁",
		},
		{
			name:             "桜都字幕组 - 摇曳露营合集",
			content:          "[桜都字幕组&7³ACG] 摇曳露营 第3季/ゆるキャン△ SEASON3/Yuru Camp S03 | 01-12+New Anime 01-03 [简繁字幕] BDrip 1080p AV1 OPUS 2.0 [复制磁连]",
			wantGroup:        "桜都字幕组&7³ACG",
			wantTitleRaw:     "ゆるキャン△",
			wantSub:          "简繁",
			wantEp:           -1,
			wantCollection:   true,
			wantEpisodeStart: 1,
			wantEpisodeEnd:   12,
		},
		{
			name:         "ANi - 碧蓝之海2",
			content:      "[ANi] Grand Blue Dreaming /  GRAND BLUE 碧蓝之海 2 - 04 [1080P][Baha][WEB-DL][AAC AVC][CHT][MP4]",
			wantGroup:    "ANi",
			wantTitleRaw: "GRAND BLUE 碧蓝之海",
			wantSeason:   2,
			wantEp:       4,
			wantSub:      "繁",
		},
		{
			name:         "萌樱字幕组 - 碧蓝之海第二季",
			content:      "[萌樱字幕组][简日双语][碧蓝之海][第二季][06][Webrip][1080p][简繁日内封]",
			wantGroup:    "萌樱字幕组",
			wantTitleRaw: "碧蓝之海",
			wantSeason:   2,
			wantEp:       6,
			wantRes:      "1080p",
			wantSub:      "简繁日",
		},
		{
			name:         "银色子弹字幕组 - 柯南",
			content:      "[银色子弹字幕组][名侦探柯南][第1071集 工藤优作的推理秀（前篇）][简日双语MP4][1080P]",
			wantGroup:    "银色子弹字幕组",
			wantTitleRaw: "名侦探柯南",
			wantEp:       1071,
			wantSub:      "简日",
		},
		{
			name:         "全遮版 - NUKITASHI",
			content:      "[全遮版&修正版&无修版] NUKITASHI住在拔作岛上的我该如何是好？ - EP06 [简／繁] (1080p&720p H.264 AAC SRTx2) {住在拔作岛上的我该如何是好？ | ぬきたし THE ANIMATION} [复制磁连]",
			wantGroup:    "全遮版&修正版&无修版",
			wantTitleRaw: "NUKITASHI住在拔作岛上的我该如何是好？ -",
			wantSub:      "简繁",
		},
		{
			name:         "DMG - FAIRY_TAIL",
			content:      "[DMG][FAIRY_TAIL - 100_YEARS_QUEST][08][1080P][GB].mp4",
			wantGroup:    "DMG",
			wantTitleRaw: "FAIRY_TAIL - 100_YEARS_QUEST",
			wantRes:      "1080P",
			wantEp:       8,
			wantSeason:   1,
			wantSub:      "简",
		},
		{
			name:         "YMDR - 哥布林杀手",
			content:      "[YMDR][哥布林杀手][Goblin Slayer][2018][02][1080p][AVC][JAP][BIG5][MP4-AAC][繁中] [复制磁连]",
			wantGroup:    "YMDR",
			wantTitleRaw: "哥布林杀手",
			wantRes:      "1080p",
			wantEp:       2,
			wantSeason:   1,
			wantSub:      "繁",
		},
		{
			name:         "安達與島村",
			content:      "【安達與島村】【第二話】【1080P】【繁體中文】【AVC】",
			wantTitleRaw: "安達與島村",
			wantRes:      "1080P",
			wantEp:       2,
			wantSeason:   1,
			wantSub:      "繁",
		},
		{
			name:           "喵萌Production&LoliHouse - U149 v2版本",
			content:        "[喵萌Production&LoliHouse] 偶像大师 灰姑娘女孩 U149 / THE IDOLM@STER CINDERELLA GIRLS U149 - 04v2 [WebRip 1080p HEVC-10bit AAC][简繁日内封字幕]",
			wantGroup:      "喵萌Production&LoliHouse",
			wantTitleRaw:   "偶像大师 灰姑娘女孩 U149",
			wantRes:        "1080p",
			wantEp:         4,
			wantSeason:     1,
			wantSub:        "简繁日",
			wantCollection: false,
			wantVersion:    2,
		},
		// ============ 合集测试用例 ============
		{
			name:             "安達與島村 - 第01-02話合集",
			content:          "【安達與島村】【第01-02話】【1080P】【繁體中文】【AVC】",
			wantTitleRaw:     "安達與島村",
			wantRes:          "1080P",
			wantEp:           -1,
			wantSub:          "繁",
			wantCollection:   true,
			wantEpisodeStart: 1,
			wantEpisodeEnd:   2,
		},
		{
			name:             "从零开始的异世界生活 - 34-35合集",
			content:          "[从零开始的异世界生活 第二季_Re Zero S2][34-35][繁体][720P][MP4]",
			wantTitleRaw:     "_Re Zero S2",
			wantRes:          "720P",
			wantSeason:       2,
			wantEp:           -1,
			wantSub:          "繁",
			wantCollection:   true,
			wantEpisodeStart: 34,
			wantEpisodeEnd:   35,
		},
		{
			name:             "Strike The Blood IV - OVA 05-06合集",
			content:          "[Strike The Blood IV][OVA][05-06][1080P][GB][MP4]",
			wantTitleRaw:     "Strike The Blood",
			wantRes:          "1080P",
			wantEp:           -1,
			wantSub:          "简",
			wantCollection:   true,
			wantSeason:       4,
			wantEpisodeStart: 5,
			wantEpisodeEnd:   6,
		},
		{
			name:           "银河英雄传说 - 全110话",
			content:        "[Legend of the Galactic Heroes 银河英雄传说][全110话+外传+剧场版][MKV][外挂繁中]",
			wantTitleRaw:   "银河英雄传说",
			wantEp:         -1,
			wantSub:        "繁",
			wantCollection: true,
		},
		{
			name:             "幻樱字幕组 - 青梅竹马 01~12合集",
			content:          "【幻樱字幕组】【青梅竹马绝对不会输的恋爱喜剧 Osamake】【01~12】【BIG5_MP4】【1280X720】【合集】",
			wantGroup:        "幻樱字幕组",
			wantTitleRaw:     "青梅竹马绝对不会输的恋爱喜剧",
			wantRes:          "1280X720",
			wantEp:           -1,
			wantSub:          "繁",
			wantCollection:   true,
			wantEpisodeStart: 1,
			wantEpisodeEnd:   12,
		},
		{
			name:             "喵萌奶茶屋 - Vivy 01-13END合集",
			content:          "【喵萌奶茶屋】★04月新番★[Vivy -Fluorite Eye's Song-][01-13END][720p][简体][招募翻译校对]",
			wantGroup:        "喵萌奶茶屋",
			wantTitleRaw:     "Vivy -Fluorite Eye's Song-",
			wantRes:          "720p",
			wantEp:           -1,
			wantSub:          "简",
			wantCollection:   true,
			wantEpisodeStart: 1,
			wantEpisodeEnd:   13,
		},
		{
			name:             "DBD-Raws - 龙珠Z 75-107合集",
			content:          "[DBD-Raws][龙珠Z 30周年纪念版/Dragonball Z 30th Anniversary Collection/ドラゴンボール Z][S3][75-107+特典映像][1080P][BDRip][HEVC-10bit][THD+AC3][MKV]",
			wantGroup:        "DBD-Raws",
			wantRes:          "1080P",
			wantSeason:       3,
			wantEp:           -1,
			wantCollection:   true,
			wantEpisodeStart: 75,
			wantEpisodeEnd:   107,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewTitleMetaParse()
			info := p.Parse(tt.content)
			if info == nil {
				t.Fatal("parser returned nil")
			}

			if tt.wantGroup != "" && info.Group != tt.wantGroup {
				t.Errorf("Group = %v, want %v", info.Group, tt.wantGroup)
			}
			if tt.wantTitleRaw != "" && info.Title != tt.wantTitleRaw {
				t.Errorf("Title = %v, want %v", info.Title, tt.wantTitleRaw)
			}
			if tt.wantRes != "" && info.Resolution != tt.wantRes {
				t.Errorf("Resolution = %v, want %v", info.Resolution, tt.wantRes)
			}
			if tt.wantEp != 0 && info.Episode != tt.wantEp {
				t.Errorf("Episode = %v, want %v", info.Episode, tt.wantEp)
			}
			if tt.wantSeason != 0 && info.Season != tt.wantSeason {
				t.Errorf("Season = %v, want %v", info.Season, tt.wantSeason)
			}
			if tt.wantSub != "" && info.Sub != tt.wantSub {
				t.Errorf("Sub = %v, want %v", info.Sub, tt.wantSub)
			}
			if info.Collection != tt.wantCollection {
				t.Errorf("Collection = %v, want %v", info.Collection, tt.wantCollection)
			}
			if tt.wantVersion != 0 && info.Version != tt.wantVersion {
				t.Errorf("Version = %v, want %v", info.Version, tt.wantVersion)
			}
			if tt.wantEpisodeStart != 0 && info.EpisodeStart != tt.wantEpisodeStart {
				t.Errorf("EpisodeStart = %v, want %v", info.EpisodeStart, tt.wantEpisodeStart)
			}
			if tt.wantEpisodeEnd != 0 && info.EpisodeEnd != tt.wantEpisodeEnd {
				t.Errorf("EpisodeEnd = %v, want %v", info.EpisodeEnd, tt.wantEpisodeEnd)
			}
		})
	}
}

func TestIsPoint5(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{
			name:    "不是0.5集",
			content: "[LoliHouse] 2.5次元的诱惑 / 2.5-jigen no Ririsa [01-24 合集][WebRip 1080p HEVC-10bit AAC][简繁内封字幕][Fin] [复制磁连]",
			want:    false,
		},
		{
			name:    "是0.5集",
			content: "[LoliHouse] 关于我转生变成史莱姆这档事 第三季 / Tensei Shitara Slime Datta Ken 3rd Season - 17.5(65.5) [WebRip 1080p HEVC-10bit AAC][简繁内封字幕] [复制磁连]",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPoint5(tt.content); got != tt.want {
				t.Errorf("IsPoint5() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsV1(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{
			name:    "不是V1版本",
			content: "[桜都字幕组&7³ACG] 摇曳露营 第3季/ゆるキャン△ SEASON3/Yuru Camp S03 | 01-12+New Anime 01-03 [简繁字幕] BDrip 1080p AV1 OPUS 2.0 [复制磁连]",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsV1(tt.content); got != tt.want {
				t.Errorf("IsV1() = %v, want %v", got, tt.want)
			}
		})
	}
}
