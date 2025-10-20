package model

// QBTorrentFile qBittorrent 种子文件信息
// 对应 API: /api/v2/torrents/files
type QBTorrentFile struct {
	Availability float64 `json:"availability"` // 可用性
	Index        int     `json:"index"`        // 文件索引
	IsSeed       bool    `json:"is_seed"`      // 是否做种
	Name         string  `json:"name"`         // 文件名（包含路径）
	PieceRange   []int   `json:"piece_range"`  // 分片范围 [开始, 结束]
	Priority     int     `json:"priority"`     // 优先级（0=不下载, 1=正常, 6=高, 7=最高）
	Progress     float64 `json:"progress"`     // 下载进度 (0-1)
	Size         int64   `json:"size"`         // 文件大小（字节）
}

// QBTorrentProperties qBittorrent 种子属性详情
// 对应 API: /api/v2/torrents/properties
type QBTorrentProperties struct {
	AdditionDate           int64   `json:"addition_date"`            // 添加时间（Unix时间戳）
	Comment                string  `json:"comment"`                  // 种子备注
	CompletionDate         int64   `json:"completion_date"`          // 完成时间（Unix时间戳，-1表示未完成）
	CreatedBy              string  `json:"created_by"`               // 创建者
	CreationDate           int64   `json:"creation_date"`            // 创建时间（Unix时间戳）
	DlLimit                int64   `json:"dl_limit"`                 // 下载限速（字节/秒，-1表示无限制）
	DlSpeed                int64   `json:"dl_speed"`                 // 下载速度（字节/秒）
	DlSpeedAvg             int64   `json:"dl_speed_avg"`             // 平均下载速度（字节/秒）
	Eta                    int64   `json:"eta"`                      // 预计剩余时间（秒，8640000表示无穷大）
	LastSeen               int64   `json:"last_seen"`                // 最后连接时间（Unix时间戳）
	NbConnections          int     `json:"nb_connections"`           // 连接数
	NbConnectionsLimit     int     `json:"nb_connections_limit"`     // 连接数限制
	Peers                  int     `json:"peers"`                    // 节点数
	PeersTotal             int     `json:"peers_total"`              // 总节点数
	PieceSize              int64   `json:"piece_size"`               // 分片大小（字节）
	PiecesHave             int     `json:"pieces_have"`              // 已有分片数
	PiecesNum              int     `json:"pieces_num"`               // 总分片数
	Reannounce             int     `json:"reannounce"`               // 重新通告时间（秒）
	SavePath               string  `json:"save_path"`                // 保存路径
	SeedingTime            int64   `json:"seeding_time"`             // 做种时间（秒）
	Seeds                  int     `json:"seeds"`                    // 种子数
	SeedsTotal             int     `json:"seeds_total"`              // 总种子数
	ShareRatio             float64 `json:"share_ratio"`              // 分享率
	TimeElapsed            int64   `json:"time_elapsed"`             // 已用时间（秒）
	TotalDownloaded        int64   `json:"total_downloaded"`         // 总下载量（字节）
	TotalDownloadedSession int64   `json:"total_downloaded_session"` // 本次会话下载量（字节）
	TotalSize              int64   `json:"total_size"`               // 总大小（字节）
	TotalUploaded          int64   `json:"total_uploaded"`           // 总上传量（字节）
	TotalUploadedSession   int64   `json:"total_uploaded_session"`   // 本次会话上传量（字节）
	TotalWasted            int64   `json:"total_wasted"`             // 浪费的数据量（字节）
	UpLimit                int64   `json:"up_limit"`                 // 上传限速（字节/秒，-1表示无限制）
	UpSpeed                int64   `json:"up_speed"`                 // 上传速度（字节/秒）
	UpSpeedAvg             int64   `json:"up_speed_avg"`             // 平均上传速度（字节/秒）
}

// QBTorrentInfo qBittorrent 种子信息
// 对应 API: /api/v2/torrents/info
type QBTorrentInfo struct {
	AddedOn           int64   `json:"added_on"`           // 添加时间（Unix时间戳）
	AmountLeft        int64   `json:"amount_left"`        // 剩余大小（字节）
	AutoTmm           bool    `json:"auto_tmm"`           // 自动种子管理
	Availability      float64 `json:"availability"`       // 可用性
	Category          string  `json:"category"`           // 分类
	Completed         int64   `json:"completed"`          // 已完成大小（字节）
	CompletionOn      int64   `json:"completion_on"`      // 完成时间（Unix时间戳，-1表示未完成）
	ContentPath       string  `json:"content_path"`       // 内容路径
	DlLimit           int64   `json:"dl_limit"`           // 下载限速（字节/秒）
	Dlspeed           int64   `json:"dlspeed"`            // 下载速度（字节/秒）
	Downloaded        int64   `json:"downloaded"`         // 已下载大小（字节）
	DownloadedSession int64   `json:"downloaded_session"` // 本次会话已下载大小（字节）
	Eta               int64   `json:"eta"`                // 预计剩余时间（秒）
	FlPiecePrio       bool    `json:"f_l_piece_prio"`     // 首尾分片优先
	ForceStart        bool    `json:"force_start"`        // 强制开始
	Hash              string  `json:"hash"`               // 种子哈希值
	LastActivity      int64   `json:"last_activity"`      // 最后活动时间（Unix时间戳）
	MagnetUri         string  `json:"magnet_uri"`         // 磁力链接
	MaxRatio          float64 `json:"max_ratio"`          // 最大分享率
	MaxSeedingTime    int64   `json:"max_seeding_time"`   // 最大做种时间（分钟）
	Name              string  `json:"name"`               // 种子名称
	NumComplete       int     `json:"num_complete"`       // 完整种子数
	NumIncomplete     int     `json:"num_incomplete"`     // 不完整种子数
	NumLeechs         int     `json:"num_leechs"`         // 下载者数量
	NumSeeds          int     `json:"num_seeds"`          // 做种者数量
	Priority          int     `json:"priority"`           // 优先级
	Progress          float64 `json:"progress"`           // 进度 (0-1)
	Ratio             float64 `json:"ratio"`              // 分享率
	RatioLimit        float64 `json:"ratio_limit"`        // 分享率限制
	SavePath          string  `json:"save_path"`          // 保存路径
	SeedingTime       int64   `json:"seeding_time"`       // 做种时间（秒）
	SeedingTimeLimit  int64   `json:"seeding_time_limit"` // 做种时间限制（分钟）
	SeenComplete      int64   `json:"seen_complete"`      // 看到完整时间（Unix时间戳）
	SeqDl             bool    `json:"seq_dl"`             // 顺序下载
	Size              int64   `json:"size"`               // 总大小（字节）
	State             string  `json:"state"`              // 状态（downloading, uploading, pausedDL, pausedUP, queuedDL, queuedUP, checkingDL, checkingUP, etc.）
	SuperSeeding      bool    `json:"super_seeding"`      // 超级做种
	Tags              string  `json:"tags"`               // 标签（逗号分隔）
	TimeActive        int64   `json:"time_active"`        // 活动时间（秒）
	TotalSize         int64   `json:"total_size"`         // 总大小（字节）
	Tracker           string  `json:"tracker"`            // 当前Tracker
	UpLimit           int64   `json:"up_limit"`           // 上传限速（字节/秒）
	Uploaded          int64   `json:"uploaded"`           // 已上传大小（字节）
	UploadedSession   int64   `json:"uploaded_session"`   // 本次会话已上传大小（字节）
	Upspeed           int64   `json:"upspeed"`            // 上传速度（字节/秒）
}
