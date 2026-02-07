package downloader

import "goto-bangumi/internal/model"

var MockTorrentInfos = map[string]*model.TorrentDownloadInfo{
	"1317e47882474c771e29ed2271b282fbfb56e7d2": {
		ETA:       0,
		SavePath:  "我推的孩子/Season 2",
		Completed: 1,
	},
	"e0a951e431269be7b556101447fbdf9d0842d72f": {
		ETA:       0,
		SavePath:  "与游戏中心的少女异文化交流的故事/Season 1",
		Completed: 1,
	},
}

var MockFiles = map[string][]string{
	"1317e47882474c771e29ed2271b282fbfb56e7d2": {"[Dynamis One] [Oshi no Ko] - 26 (ABEMA 1920x1080 AVC AAC MP4) [8DF340A3].mp4"},
	"e0a951e431269be7b556101447fbdf9d0842d72f": {
		"[三明治摆烂组&Prejudice-Studio] 与游戏中心的少女异文化交流的故事 [01-12 合集][简日内嵌][H264 8bit 1080P]/与游戏中心的少女异文化交流的故事 - S01E06 - [三明治摆烂组&Prejudice-Studio][简日内嵌][H264 8bit 1080P].mp4",
		"[三明治摆烂组&Prejudice-Studio] 与游戏中心的少女异文化交流的故事 [01-12 合集][简日内嵌][H264 8bit 1080P]/与游戏中心的少女异文化交流的故事 - S01E02 - [三明治摆烂组&Prejudice-Studio][简日内嵌][H264 8bit 1080P].mp4",
		"[三明治摆烂组&Prejudice-Studio] 与游戏中心的少女异文化交流的故事 [01-12 合集][简日内嵌][H264 8bit 1080P]/与游戏中心的少女异文化交流的故事 - S01E03 - [三明治摆烂组&Prejudice-Studio][简日内嵌][H264 8bit 1080P].mp4",
		"[三明治摆烂组&Prejudice-Studio] 与游戏中心的少女异文化交流的故事 [01-12 合集][简日内嵌][H264 8bit 1080P]/与游戏中心的少女异文化交流的故事 - S01E04 - [三明治摆烂组&Prejudice-Studio][简日内嵌][H264 8bit 1080P].mp4",
		"[三明治摆烂组&Prejudice-Studio] 与游戏中心的少女异文化交流的故事 [01-12 合集][简日内嵌][H264 8bit 1080P]/与游戏中心的少女异文化交流的故事 - S01E05 - [三明治摆烂组&Prejudice-Studio][简日内嵌][H264 8bit 1080P].mp4",
		"[三明治摆烂组&Prejudice-Studio] 与游戏中心的少女异文化交流的故事 [01-12 合集][简日内嵌][H264 8bit 1080P]/与游戏中心的少女异文化交流的故事 - S01E01 - [三明治摆烂组&Prejudice-Studio][简日内嵌][H264 8bit 1080P].mp4",
		"[三明治摆烂组&Prejudice-Studio] 与游戏中心的少女异文化交流的故事 [01-12 合集][简日内嵌][H264 8bit 1080P]/与游戏中心的少女异文化交流的故事 - S01E07 - [三明治摆烂组&Prejudice-Studio][简日内嵌][H264 8bit 1080P].mp4",
		"[三明治摆烂组&Prejudice-Studio] 与游戏中心的少女异文化交流的故事 [01-12 合集][简日内嵌][H264 8bit 1080P]/与游戏中心的少女异文化交流的故事 - S01E08 - [三明治摆烂组&Prejudice-Studio][简日内嵌][H264 8bit 1080P].mp4",
		"[三明治摆烂组&Prejudice-Studio] 与游戏中心的少女异文化交流的故事 [01-12 合集][简日内嵌][H264 8bit 1080P]/与游戏中心的少女异文化交流的故事 - S01E09 - [三明治摆烂组&Prejudice-Studio][简日内嵌][H264 8bit 1080P].mp4",
		"[三明治摆烂组&Prejudice-Studio] 与游戏中心的少女异文化交流的故事 [01-12 合集][简日内嵌][H264 8bit 1080P]/与游戏中心的少女异文化交流的故事 - S01E10 - [三明治摆烂组&Prejudice-Studio][简日内嵌][H264 8bit 1080P].mp4",
		"[三明治摆烂组&Prejudice-Studio] 与游戏中心的少女异文化交流的故事 [01-12 合集][简日内嵌][H264 8bit 1080P]/与游戏中心的少女异文化交流的故事 - S01E11 - [三明治摆烂组&Prejudice-Studio][简日内嵌][H264 8bit 1080P].mp4",
		"[三明治摆烂组&Prejudice-Studio] 与游戏中心的少女异文化交流的故事 [01-12 合集][简日内嵌][H264 8bit 1080P]/与游戏中心的少女异文化交流的故事 - S01E12 - [三明治摆烂组&Prejudice-Studio][简日内嵌][H264 8bit 1080P].mp4",
	},
}
