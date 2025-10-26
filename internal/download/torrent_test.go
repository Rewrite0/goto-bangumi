package download

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseTorrent(t *testing.T) {
	// 定义测试用例
	tests := []struct {
		name         string
		filePath     string
		wantName     string
		wantHashV1   string
		wantHashV2   string
		wantErr      bool
	}{
		{
			name:       "Hybrid torrent (V1 + V2)",
			filePath:   "./test_data/test1.torrent",
			wantName:   "[樱桃花字幕组] Rock wa Lady no Tashinami deshite - 02（1080P） [B7ADB92B].mp4",
			wantHashV1: "32587df888ce2b3f7d9df67854ea10e50153a55c",
			wantHashV2: "17aec1ef77c038f5164ce6a8d4d40cd3674b325ec663ac5705a0c5d697a26b39",
			wantErr:    false,
		},
		{
			name:       "V1 only torrent",
			filePath:   "./test_data/test2.torrent",
			wantName:   "[Skymoon-Raws] SPY×FAMILY Season 3 - 41 [ViuTV][WEB-DL][CHT][SRT][1080p][AVC AAC].mkv",
			wantHashV1: "7a34f9ba65b362c424524057882357e191368f2e",
			wantHashV2: "",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 读取种子文件
			absPath, err := filepath.Abs(tt.filePath)
			if err != nil {
				t.Fatalf("Failed to get absolute path: %v", err)
			}

			data, err := os.ReadFile(absPath)
			if err != nil {
				t.Fatalf("Failed to read torrent file %s: %v", tt.filePath, err)
			}

			// 解析种子
			torrentInfo, err := ParseTorrent(data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTorrent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				// 验证解析结果
				if torrentInfo.Name != tt.wantName {
					t.Errorf("Name = %v, want %v", torrentInfo.Name, tt.wantName)
				}
				if torrentInfo.InfoHashV1 != tt.wantHashV1 {
					t.Errorf("InfoHashV1 = %v, want %v", torrentInfo.InfoHashV1, tt.wantHashV1)
				}
				if torrentInfo.InfoHashV2 != tt.wantHashV2 {
					t.Errorf("InfoHashV2 = %v, want %v", torrentInfo.InfoHashV2, tt.wantHashV2)
				}

				t.Logf("✓ Torrent Info:\n%s", torrentInfo.String())
			}
		})
	}
}
