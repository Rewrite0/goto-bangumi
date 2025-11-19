package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *slog.Logger

var (
	logDir  = "./data"
	logFile = "log.txt"
)

// CustomHandler 自定义日志处理器
type CustomHandler struct {
	writer io.Writer
	level  slog.Level
	attrs  []slog.Attr
	groups []string
}

// Enabled 判断是否启用该日志级别
func (h *CustomHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

// Handle 处理日志记录
func (h *CustomHandler) Handle(_ context.Context, r slog.Record) error {
	// 格式: [2025-11-15 19:13:04] DEBUG:   [module] message
	buf := &strings.Builder{}

	// 时间戳: [2025-11-15 19:13:04]
	buf.WriteString("[")
	buf.WriteString(r.Time.Format("2006-01-02 15:04:05"))
	buf.WriteString("] ")

	// 日志级别: DEBUG:
	buf.WriteString(r.Level.String())
	buf.WriteString(":")

	// 补齐空格使对齐
	levelPadding := 7 - len(r.Level.String())
	for i := 0; i < levelPadding; i++ {
		buf.WriteString(" ")
	}

	var otherAttrs []string

	// 再处理记录级别的属性
	r.Attrs(func(attr slog.Attr) bool {
		// 格式化其他属性
		otherAttrs = append(otherAttrs, fmt.Sprintf("%s=%v", attr.Key, attr.Value.Any()))
		return true
	})

	// 日志消息
	buf.WriteString(r.Message)

	// 其他属性
	if len(otherAttrs) > 0 {
		buf.WriteString(" ")
		buf.WriteString(strings.Join(otherAttrs, " "))
	}

	buf.WriteString("\n")

	_, err := h.writer.Write([]byte(buf.String()))
	return err
}

// WithAttrs 返回带有额外属性的新处理器
func (h *CustomHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)

	return &CustomHandler{
		writer: h.writer,
		level:  h.level,
		attrs:  newAttrs,
		groups: h.groups,
	}
}

// WithGroup 返回带有组名的新处理器
func (h *CustomHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = name

	return &CustomHandler{
		writer: h.writer,
		level:  h.level,
		attrs:  h.attrs,
		groups: newGroups,
	}
}

func init() {
	// 检测配置文件夹是否存在, 不存在则创建
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		if err := os.MkdirAll(logDir, 0o755); err != nil {
			slog.Error("创建配置文件夹失败", "error", err)
			return
		}
		slog.Info("配置文件夹创建成功", "path", logDir)
	}
}

// Init 初始化日志系统
// debugEnable: 是否开启调试模式，true 为 Debug 级别，false 为 Info 级别
func Init(debugEnable bool) {
	var level slog.Level
	if debugEnable {
		level = slog.LevelDebug
	} else {
		level = slog.LevelInfo
	}

	// 确保日志目录存在
	logDir := "./data"
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		panic("创建日志目录失败: " + err.Error())
	}

	// 配置 lumberjack 进行日志轮转
	fileWriter := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, logFile), // 日志文件路径
		MaxSize:    10,                             // 单个文件最大 10MB
		MaxBackups: 3,                              // 最多保留 3 个旧文件
		MaxAge:     14,                             // 保留 14 天
		Compress:   false,                          // 不压缩旧文件
	}

	// 同时输出到控制台和文件
	multiWriter := io.MultiWriter(os.Stdout, fileWriter)

	// 创建自定义格式的日志处理器
	handler := &CustomHandler{
		writer: multiWriter,
		level:  level,
		attrs:  []slog.Attr{},
		groups: []string{},
	}

	Logger = slog.New(handler)

	// 设置为默认日志记录器
	slog.SetDefault(Logger)

	Logger.Info("日志系统已初始化", "level", level.String(), "file", filepath.Join(logDir, "log.txt"))
}

// GetLogger 获取全局日志实例
func GetLogger() *slog.Logger {
	if Logger == nil {
		// 如果未初始化，使用默认配置
		Init(false)
	}
	return Logger
}
