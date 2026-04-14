package rename

import (
	"context"
	"fmt"
	"testing"

	"goto-bangumi/internal/download"
	"goto-bangumi/internal/download/downloader"
	"goto-bangumi/internal/model"
)

// setupMockClient 创建用于测试的 mock download client
func setupMockClient() *download.DownloadClient {
	mockDownloader := downloader.NewMockDownloader()
	mockConfig := &model.DownloaderConfig{
		SavePath: "",
		Type:     "mock",
	}
	mockDownloader.Init(mockConfig)

	dlClient := download.NewDownloadClient()
	dlClient.Init(mockConfig)
	dlClient.Downloader = mockDownloader
	return dlClient
}

func TestRename_SingleFile(t *testing.T) {
	// 我推的孩子 Season 2, 单个文件
	// 种子名: [Dynamis One] [Oshi no Ko] - 26 (ABEMA 1920x1080 AVC AAC MP4) [8DF340A3].mp4
	Init(&model.BangumiRenameConfig{
		Year:  false,
		Group: false,
	})

	dlClient := setupMockClient()
	r := New(nil, dlClient)

	torrent := &model.Torrent{
		DownloadUID: "1317e47882474c771e29ed2271b282fbfb56e7d2",
		Name:        "[Dynamis One] [Oshi no Ko] - 26 (ABEMA 1920x1080 AVC AAC MP4) [8DF340A3].mp4",
	}
	bangumi := &model.Bangumi{
		OfficialTitle: "我推的孩子",
		Season:        2,
	}

	ctx := context.Background()
	r.Rename(ctx, torrent, bangumi)

	// 验证文件是否已重命名
	files, err := dlClient.GetTorrentFiles(ctx, torrent.DownloadUID)
	if err != nil {
		t.Fatalf("GetTorrentFiles() error = %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	want := "我推的孩子 S02E26.mp4"
	if files[0] != want {
		t.Errorf("renamed file = %q, want %q", files[0], want)
	}
}

func TestRename_SingleFileWithYearAndGroup(t *testing.T) {
	// 同一个种子,但开启年份和字幕组
	Init(&model.BangumiRenameConfig{
		Year:  true,
		Group: true,
	})

	dlClient := setupMockClient()
	r := New(nil, dlClient)

	torrent := &model.Torrent{
		DownloadUID: "1317e47882474c771e29ed2271b282fbfb56e7d2",
		Name:        "[Dynamis One] [Oshi no Ko] - 26 (ABEMA 1920x1080 AVC AAC MP4) [8DF340A3].mp4",
	}
	bangumi := &model.Bangumi{
		OfficialTitle: "我推的孩子",
		Season:        2,
		Year:          "2024",
	}

	ctx := context.Background()
	r.Rename(ctx, torrent, bangumi)

	files, err := dlClient.GetTorrentFiles(ctx, torrent.DownloadUID)
	if err != nil {
		t.Fatalf("GetTorrentFiles() error = %v", err)
	}

	// Dynamis One 不在已知字幕组列表中, parser 不会提取 group
	want := "我推的孩子 (2024) S02E26.mp4"
	if files[0] != want {
		t.Errorf("renamed file = %q, want %q", files[0], want)
	}
}

func TestRename_MultipleFiles(t *testing.T) {
	// 与游戏中心的少女异文化交流的故事, 12集合集
	// 种子文件路径带目录前缀,Rename 应该取 filepath.Base 再生成新路径
	Init(&model.BangumiRenameConfig{
		Year:  false,
		Group: false,
	})

	dlClient := setupMockClient()
	r := New(nil, dlClient)

	torrent := &model.Torrent{
		DownloadUID: "e0a951e431269be7b556101447fbdf9d0842d72f",
		Name:        "[三明治摆烂组&Prejudice-Studio] 与游戏中心的少女异文化交流的故事 [01-12 合集]",
	}
	bangumi := &model.Bangumi{
		OfficialTitle: "与游戏中心的少女异文化交流的故事",
		Season:        1,
	}

	ctx := context.Background()
	r.Rename(ctx, torrent, bangumi)

	files, err := dlClient.GetTorrentFiles(ctx, torrent.DownloadUID)
	if err != nil {
		t.Fatalf("GetTorrentFiles() error = %v", err)
	}

	if len(files) != 12 {
		t.Fatalf("expected 12 files, got %d", len(files))
	}

	// 验证 12 集都存在且格式正确
	seen := make(map[string]bool)
	for _, f := range files {
		seen[f] = true
	}
	for ep := 1; ep <= 12; ep++ {
		want := fmt.Sprintf("与游戏中心的少女异文化交流的故事 S01E%02d.mp4", ep)
		if !seen[want] {
			t.Errorf("missing expected file %q", want)
		}
	}
}

func TestRename_WithOffset(t *testing.T) {
	// 模拟第二季偏移量场景: 种子里集数是14, 偏移-12 后变成 E02
	Init(&model.BangumiRenameConfig{
		Year:  false,
		Group: false,
	})

	// 手动设置 mock 数据
	mockDownloader := downloader.NewMockDownloader()
	mockConfig := &model.DownloaderConfig{
		SavePath: "",
		Type:     "mock",
	}
	mockDownloader.Init(mockConfig)

	dlClient := download.NewDownloadClient()
	dlClient.Init(mockConfig)
	dlClient.Downloader = mockDownloader

	// 通过 Add 方法添加自定义种子
	hash := "abc123test"
	mockDownloader.AddMockTorrent(hash, &model.TorrentDownloadInfo{
		SavePath:  "转生贵族靠鉴定技能一飞冲天/Season 2",
		Completed: 1,
	}, []string{
		"[ANi] 转生贵族靠鉴定技能一飞冲天 - 14 [1080P][Baha][WEB-DL][AAC AVC][CHT].mp4",
	})

	r := New(nil, dlClient)
	torrent := &model.Torrent{
		DownloadUID: hash,
		Name:        "[ANi] 转生贵族靠鉴定技能一飞冲天 - 14 [1080P][Baha][WEB-DL][AAC AVC][CHT].mp4",
	}
	bangumi := &model.Bangumi{
		OfficialTitle: "转生贵族靠鉴定技能一飞冲天",
		Season:        2,
		Offset:        -12,
	}

	ctx := context.Background()
	r.Rename(ctx, torrent, bangumi)

	files, err := dlClient.GetTorrentFiles(ctx, hash)
	if err != nil {
		t.Fatalf("GetTorrentFiles() error = %v", err)
	}

	want := "转生贵族靠鉴定技能一飞冲天 S02E02.mp4"
	if len(files) != 1 || files[0] != want {
		t.Errorf("renamed file = %q, want %q", files, want)
	}
}

func TestRename_NilBangumi_FallbackToGetBangumi(t *testing.T) {
	// bangumi 为 nil 时, Rename 会调用 getBangumi 从路径解析
	Init(&model.BangumiRenameConfig{
		Year:  false,
		Group: false,
	})

	dlClient := setupMockClient()
	r := New(nil, dlClient)

	torrent := &model.Torrent{
		DownloadUID: "1317e47882474c771e29ed2271b282fbfb56e7d2",
		Name:        "[Dynamis One] [Oshi no Ko] - 26 (ABEMA 1920x1080 AVC AAC MP4) [8DF340A3].mp4",
	}

	ctx := context.Background()
	// bangumi 传 nil, 应自动从 savePath "我推的孩子/Season 2" 解析出来
	r.Rename(ctx, torrent, nil)

	files, err := dlClient.GetTorrentFiles(ctx, torrent.DownloadUID)
	if err != nil {
		t.Fatalf("GetTorrentFiles() error = %v", err)
	}

	want := "我推的孩子 S02E26.mp4"
	if len(files) != 1 || files[0] != want {
		t.Errorf("renamed file = %q, want %q", files[0], want)
	}
}

func TestRename_SkipSameFilename(t *testing.T) {
	// 如果新路径和旧路径相同, 应该跳过重命名
	Init(&model.BangumiRenameConfig{
		Year:  false,
		Group: false,
	})

	mockDownloader := downloader.NewMockDownloader()
	mockConfig := &model.DownloaderConfig{
		SavePath: "",
		Type:     "mock",
	}
	mockDownloader.Init(mockConfig)

	dlClient := download.NewDownloadClient()
	dlClient.Init(mockConfig)
	dlClient.Downloader = mockDownloader

	// 文件名已经是目标格式
	alreadyRenamed := "败犬女主太多了 S01E02.mp4"
	hash := "skip_same_test"
	mockDownloader.AddMockTorrent(hash, &model.TorrentDownloadInfo{
		SavePath:  "败犬女主太多了/Season 1",
		Completed: 1,
	}, []string{alreadyRenamed})

	r := New(nil, dlClient)
	torrent := &model.Torrent{
		DownloadUID: hash,
		Name:        alreadyRenamed,
	}
	bangumi := &model.Bangumi{
		OfficialTitle: "败犬女主太多了",
		Season:        1,
	}

	ctx := context.Background()
	r.Rename(ctx, torrent, bangumi)

	files, err := dlClient.GetTorrentFiles(ctx, hash)
	if err != nil {
		t.Fatalf("GetTorrentFiles() error = %v", err)
	}

	// 文件名应保持不变
	if files[0] != alreadyRenamed {
		t.Errorf("file should not be renamed, got %q", files[0])
	}
}

func TestRename_MixedSubGroups(t *testing.T) {
	// 测试不同字幕组格式的种子重命名
	Init(&model.BangumiRenameConfig{
		Year:  true,
		Group: true,
	})

	mockDownloader := downloader.NewMockDownloader()
	mockConfig := &model.DownloaderConfig{
		SavePath: "",
		Type:     "mock",
	}
	mockDownloader.Init(mockConfig)

	dlClient := download.NewDownloadClient()
	dlClient.Init(mockConfig)
	dlClient.Downloader = mockDownloader

	tests := []struct {
		name     string
		hash     string
		file     string
		bangumi  *model.Bangumi
		wantFile string
	}{
		{
			name: "ANi字幕组-败犬女主太多了",
			hash: "test_ani_makeine",
			file: "[ANi] 败犬女主太多了！ - 02 [1080p][Baha][WEB-DL][AAC AVC][CHT].mp4",
			bangumi: &model.Bangumi{
				OfficialTitle: "败犬女主太多了",
				Season:        1,
				Year:          "2024",
			},
			wantFile: "败犬女主太多了 (2024) S01E02 - ANi.mp4",
		},
		{
			name: "Lilith-Raws-我的英雄学院",
			hash: "test_lilith_mha",
			file: "[Lilith-Raws] 我的英雄学院 第七季 - 08 [Baha][WEB-DL][1080p].mp4",
			bangumi: &model.Bangumi{
				OfficialTitle: "我的英雄学院",
				Season:        7,
				Year:          "2024",
			},
			wantFile: "我的英雄学院 (2024) S07E08 - Lilith-Raws.mp4",
		},
		{
			// Skymoon-Raws 不在已知字幕组列表中, 但通过 fallback 取第一个 token 作为 group
			name: "Skymoon-Raws-SPY×FAMILY",
			hash: "test_skymoon_spy",
			file: "[Skymoon-Raws] SPY×FAMILY Season 3 - 41 [ViuTV][WEB-DL][CHT][SRT][1080p][AVC AAC].mkv",
			bangumi: &model.Bangumi{
				OfficialTitle: "SPY×FAMILY",
				Season:        3,
				Year:          "2025",
			},
			wantFile: "SPY×FAMILY (2025) S03E41 - Skymoon-Raws.mkv",
		},
		{
			name: "樱桃花字幕组-Rock wa Lady",
			hash: "test_sakura_rock",
			file: "[樱桃花字幕组] Rock wa Lady no Tashinami deshite - 02（1080P） [B7ADB92B].mp4",
			bangumi: &model.Bangumi{
				OfficialTitle: "淑女养成摇滚",
				Season:        1,
				Year:          "2025",
			},
			wantFile: "淑女养成摇滚 (2025) S01E02 - 樱桃花字幕组.mp4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDownloader.AddMockTorrent(tt.hash, &model.TorrentDownloadInfo{
				SavePath:  tt.bangumi.OfficialTitle + "/Season 1",
				Completed: 1,
			}, []string{tt.file})

			r := New(nil, dlClient)
			torrent := &model.Torrent{
				DownloadUID: tt.hash,
				Name:        tt.file,
			}

			ctx := context.Background()
			r.Rename(ctx, torrent, tt.bangumi)

			files, err := dlClient.GetTorrentFiles(ctx, tt.hash)
			if err != nil {
				t.Fatalf("GetTorrentFiles() error = %v", err)
			}

			if len(files) != 1 || files[0] != tt.wantFile {
				t.Errorf("renamed file = %q, want %q", files[0], tt.wantFile)
			}
		})
	}
}
