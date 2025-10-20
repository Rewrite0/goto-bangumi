package main

import (
	"fmt"
	"log"

	"goto-bangumi/internal/download"
	"goto-bangumi/internal/model"
)

func main() {
	// 创建默认配置
	config := model.NewDownloaderConfig()

	config.Host = "http://localhost:8999"
	// config.Password = "your_password"
	// 使用工厂函数创建下载器，根据 config.Type 动态选择
	d, err := download.NewDownloader(config.Type, config)
	if err != nil {
		log.Fatalf("创建下载器失败: %v", err)
	}

	// 测试认证
	success, err := d.Auth()
	if err != nil {
		log.Fatalf("认证失败: %v", err)
	}

	if success {
		fmt.Println("✓ 下载器认证成功!")
	} else {
		fmt.Println("✗ 下载器认证失败!")
	}
	// 测试 get torrent info
	hash := "c3dc7d6a37fe7c6334d38e8d2ce18fe285ff9da2"
	torrents, err := d.GetTorrentFiles(hash)
	if err != nil {
		log.Fatalf("获取种子文件列表失败: %v", err)
	}
	
	fmt.Printf("种子文件列表: %v\n", torrents)
}
