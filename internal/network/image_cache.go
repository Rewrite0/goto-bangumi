package network

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

var (
	dataDir   = "data"
	posterDir = filepath.Join(dataDir, "posters")
)

func init() {
	// 创建海报目录（如果不存在）
	if err := os.MkdirAll(posterDir, 0o755); err != nil {
		slog.Error("failed to create poster directory", "error", err)
	}
}

// urlToBase64 converts URL to base64 encoded string for filename (reversible)
func urlToBase64(url string) string {
	return base64.URLEncoding.EncodeToString([]byte(url))
}

// base64ToURL decodes base64 string back to URL
func base64ToURL(encoded string) (string, error) {
	decoded, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		// 如果不是有效的 base64，直接返回原字符串
		return "", fmt.Errorf("failed to decode base64 string: %w", err)
	}
	// 如果不是有效的 URL，返回错误
	if !strings.HasPrefix(string(decoded), "http") {
		return "", fmt.Errorf("decoded string is not a valid URL")
	}
	return string(decoded), nil
}

// SaveImage downloads and saves an image to cache
func SaveImage(url string) ([]byte, error) {
	// Generate base64 encoded filename
	imgEncoded := urlToBase64(url)
	imagePath := filepath.Join(posterDir, imgEncoded)

	// Download image
	imgData, err := defaultClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	// TODO:检测是不是图片以及是不是有效数据
	if len(imgData) == 0 {
		return nil, fmt.Errorf("downloaded image is empty")
	}

	// Save to file
	if err := os.WriteFile(imagePath, imgData, 0o644); err != nil {
		return nil, fmt.Errorf("failed to save image: %w", err)
	}

	slog.Info("[ImageCache] Saved image", "url", url, "path", imagePath)
	return imgData, nil
}

// LoadImage 从缓存加载图片，如果不存在则下载
func LoadImage(imgPath string) ([]byte, error) {
	// Check if it's a URL
	if strings.HasPrefix(imgPath, "http") {
		imgPath = urlToBase64(imgPath)
	}

	imagePath := filepath.Join(posterDir, imgPath)

	// 如果文件存在，直接读取
	if data, err := os.ReadFile(imagePath); err == nil {
		return data, nil
	}

	// 文件不存在，尝试下载
	slog.Info("[ImageCache] Image not found in cache, downloading", "path", imgPath)

	// 将 base64 解码回 URL
	decodedURL, err := base64ToURL(imgPath)
	if err != nil {
		slog.Debug("[ImageCache] Decoding as URL failed, treating as direct URL", "error", err)
		return nil, err
	}

	// 如果解码后不是有效 URL，则报错
	if !strings.HasPrefix(decodedURL, "http") {
		return nil, fmt.Errorf("cannot download image: invalid URL from path %s", imgPath)
	}

	imgData, err := SaveImage(decodedURL)
	if err != nil {
		return nil, err
	}

	return imgData, nil
}
