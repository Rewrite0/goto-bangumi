package logger

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func resetState(t *testing.T) {
	t.Helper()
	t.Chdir(t.TempDir())
	origDefault := slog.Default()
	t.Cleanup(func() {
		Close()
		slog.SetDefault(origDefault)
		Logger = nil
		logLevel.Set(slog.LevelInfo)
	})
}

func TestInit_DefaultLevelIsInfo(t *testing.T) {
	resetState(t)
	Init()

	if Logger == nil {
		t.Fatal("Logger is nil after Init")
	}
	if logLevel.Level() != slog.LevelInfo {
		t.Errorf("expected default level Info, got %v", logLevel.Level())
	}
	if Logger.Enabled(context.Background(), slog.LevelDebug) {
		t.Error("Debug should be suppressed at default Info level")
	}
	if !Logger.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("Info should be enabled at default Info level")
	}

	SetLevel(slog.LevelDebug)
	if !Logger.Enabled(context.Background(), slog.LevelDebug) {
		t.Error("Debug should pass after SetLevel(Debug)")
	}

	SetLevel(slog.LevelWarn)
	if Logger.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("Info should be filtered after SetLevel(Warn)")
	}
	if !Logger.Enabled(context.Background(), slog.LevelWarn) {
		t.Error("Warn should pass after SetLevel(Warn)")
	}
}

func TestInit_OutputFormat(t *testing.T) {
	resetState(t)

	buf := &bytes.Buffer{}
	lv := new(slog.LevelVar)
	h := &CustomHandler{writer: buf, level: lv}

	ts := time.Date(2024, 3, 15, 10, 20, 30, 0, time.UTC)
	r := slog.NewRecord(ts, slog.LevelInfo, "测试消息", 0)
	r.AddAttrs(slog.String("key", "val"))

	if err := h.Handle(context.Background(), r); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	for _, want := range []string{
		"[2024-03-15 10:20:30]",
		"INFO:",
		"测试消息",
		"key=val",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output %q missing %q", out, want)
		}
	}
	if !strings.HasSuffix(out, "\n") {
		t.Errorf("output should end with newline: %q", out)
	}
}
