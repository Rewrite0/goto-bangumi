package rename

import (
	"testing"

	"goto-bangumi/internal/model"
)

func TestGenPath(t *testing.T) {
	tests := []struct {
		name        string
		torrentName string
		bangumi     *model.Bangumi
		config      *model.BangumiRenameConfig
		wantPath    string
		wantEpisode int
	}{
		{
			name:        "基础情况-只有标题季度集数",
			torrentName: "[ANi] 败犬女主太多了！ - 02 [1080p][Baha][WEB-DL][AAC AVC][CHT].mp4",
			bangumi: &model.Bangumi{
				OfficialTitle: "败犬女主太多了",
				Season:        1,
			},
			config: &model.BangumiRenameConfig{
				Year:  false,
				Group: false,
			},
			wantPath:    "败犬女主太多了 S01E02.mp4",
			wantEpisode: 2,
		},
		{
			name:        "带年份",
			torrentName: "[Nekomoe kissaten][Makeine][02][1080p][JPSC].mp4",
			bangumi: &model.Bangumi{
				OfficialTitle: "败犬女主太多了",
				Season:        1,
				Year:          "2024",
			},
			config: &model.BangumiRenameConfig{
				Year:  true,
				Group: false,
			},
			wantPath:    "败犬女主太多了 (2024) S01E02.mp4",
			wantEpisode: 2,
		},
		{
			name:        "带字幕组",
			torrentName: "[ANi] 败犬女主太多了！ - 03 [1080p][Baha][WEB-DL][AAC AVC][CHT].mp4",
			bangumi: &model.Bangumi{
				OfficialTitle: "败犬女主太多了",
				Season:        1,
			},
			config: &model.BangumiRenameConfig{
				Year:  false,
				Group: true,
			},
			wantPath:    "败犬女主太多了 S01E03 - ANi.mp4",
			wantEpisode: 3,
		},
		{
			name:        "带年份和字幕组",
			torrentName: "[ANi] 败犬女主太多了！ - 04 [1080p][Baha][WEB-DL][AAC AVC][CHT].mp4",
			bangumi: &model.Bangumi{
				OfficialTitle: "败犬女主太多了",
				Season:        1,
				Year:          "2024",
			},
			config: &model.BangumiRenameConfig{
				Year:  true,
				Group: true,
			},
			wantPath:    "败犬女主太多了 (2024) S01E04 - ANi.mp4",
			wantEpisode: 4,
		},
		{
			name:        "带偏移量",
			torrentName: "[ANi] 败犬女主太多了！ - 02 [1080p].mp4",
			bangumi: &model.Bangumi{
				OfficialTitle: "败犬女主太多了",
				Season:        2,
				Offset:        12,
			},
			config: &model.BangumiRenameConfig{
				Year:  false,
				Group: false,
			},
			wantPath:    "败犬女主太多了 S02E14.mp4",
			wantEpisode: 2,
		},
		{
			name:        "mkv扩展名",
			torrentName: "[Nekomoe kissaten][Makeine][05][1080p][JPSC].mkv",
			bangumi: &model.Bangumi{
				OfficialTitle: "败犬女主太多了",
				Season:        1,
			},
			config: &model.BangumiRenameConfig{
				Year:  false,
				Group: false,
			},
			wantPath:    "败犬女主太多了 S01E05.mkv",
			wantEpisode: 5,
		},
		{
			name:        "两位数集数",
			torrentName: "[ANi] 海贼王 - 1120 [1080p].mp4",
			bangumi: &model.Bangumi{
				OfficialTitle: "海贼王",
				Season:        1,
			},
			config: &model.BangumiRenameConfig{
				Year:  false,
				Group: false,
			},
			wantPath:    "海贼王 S01E1120.mp4",
			wantEpisode: 1120,
		},
		{
			name:        "年份配置开启但bangumi没有年份",
			torrentName: "[ANi] 测试番剧 - 01 [1080p].mp4",
			bangumi: &model.Bangumi{
				OfficialTitle: "测试番剧",
				Season:        1,
				Year:          "",
			},
			config: &model.BangumiRenameConfig{
				Year:  true,
				Group: false,
			},
			wantPath:    "测试番剧 S01E01.mp4",
			wantEpisode: 1,
		},
		{
			name:        "多季度番剧",
			torrentName: "[Lilith-Raws] 我的英雄学院 第七季 - 08 [Baha][WEB-DL][1080p].mp4",
			bangumi: &model.Bangumi{
				OfficialTitle: "我的英雄学院",
				Season:        7,
			},
			config: &model.BangumiRenameConfig{
				Year:  false,
				Group: false,
			},
			wantPath:    "我的英雄学院 S07E08.mp4",
			wantEpisode: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置配置
			Init(tt.config)

			meta, gotPath := GenPath(tt.torrentName, tt.bangumi)

			if gotPath != tt.wantPath {
				t.Errorf("GenPath() path = %q, want %q", gotPath, tt.wantPath)
			}

			if meta != nil && meta.Episode != tt.wantEpisode {
				t.Errorf("GenPath() episode = %d, want %d", meta.Episode, tt.wantEpisode)
			}
		})
	}
}

func TestGenPath_ParseFailed(t *testing.T) {
	// 设置配置
	Init(&model.BangumiRenameConfig{
		Year:  false,
		Group: false,
	})

	// 测试无法解析集数的情况 - 合集类型会返回 -1
	torrentName := "[字幕组] 番剧名 [全集][1080p].mp4"
	bangumi := &model.Bangumi{
		OfficialTitle: "测试番剧",
		Season:        1,
	}

	meta, gotPath := GenPath(torrentName, bangumi)

	// 解析失败时应该返回空字符串
	if gotPath != "" {
		t.Errorf("GenPath() with invalid torrent name should return empty path, got %q", gotPath)
	}

	if meta != nil {
		t.Logf("meta.Episode = %d (expected -1 for collection)", meta.Episode)
	}
}

func TestGenPath_NegativeOffset(t *testing.T) {
	Init(&model.BangumiRenameConfig{
		Year:  false,
		Group: false,
	})

	torrentName := "[ANi] 转生贵族靠鉴定技能一飞冲天 - 14 [1080p].mp4"
	bangumi := &model.Bangumi{
		OfficialTitle: "转生贵族靠鉴定技能一飞冲天",
		Season:        2,
		Offset:        -12,
	}

	meta, gotPath := GenPath(torrentName, bangumi)

	wantPath := "转生贵族靠鉴定技能一飞冲天 S02E02.mp4"
	if gotPath != wantPath {
		t.Errorf("GenPath() path = %q, want %q", gotPath, wantPath)
	}

	if meta != nil && meta.Episode != 14 {
		t.Errorf("GenPath() raw episode = %d, want 14", meta.Episode)
	}
}
