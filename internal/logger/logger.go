package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	logDir  = "./data"
	logFile = "log.txt"
)

var (
	Logger   *slog.Logger
	logLevel = new(slog.LevelVar) // 默认 Info
	closer   io.Closer
)

type CustomHandler struct {
	writer io.Writer
	level  *slog.LevelVar
	mu     sync.Mutex
}

func (h *CustomHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

func (h *CustomHandler) Handle(_ context.Context, r slog.Record) error {
	buf := &strings.Builder{}

	buf.WriteString("[")
	buf.WriteString(r.Time.Format("2006-01-02 15:04:05"))
	buf.WriteString("] ")

	buf.WriteString(r.Level.String())
	buf.WriteString(":")
	buf.WriteString(strings.Repeat(" ", 7-len(r.Level.String())))

	buf.WriteString(r.Message)

	r.Attrs(func(attr slog.Attr) bool {
		fmt.Fprintf(buf, " %s=%v", attr.Key, attr.Value.Any())
		return true
	})

	buf.WriteString("\n")

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := io.WriteString(h.writer, buf.String())
	return err
}

func (h *CustomHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *CustomHandler) WithGroup(_ string) slog.Handler      { return h }

func Init() {
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		panic("创建日志目录失败: " + err.Error())
	}

	fileWriter := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, logFile),
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     14,
	}
	closer = fileWriter

	Logger = slog.New(&CustomHandler{
		writer: io.MultiWriter(os.Stdout, fileWriter),
		level:  logLevel,
	})
	slog.SetDefault(Logger)

	Logger.Info("日志系统已初始化", "level", logLevel.Level().String(), "file", filepath.Join(logDir, logFile))
}

func Close() {
	if closer != nil {
		closer.Close()
		closer = nil
	}
}

func SetLevel(level slog.Level) {
	logLevel.Set(level)
}

func GetLogger() *slog.Logger {
	if Logger == nil {
		Init()
	}
	return Logger
}
