# Notification System Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Simplify the notification system so it only sends pre-formatted messages, fix the bug where Send never actually sends, and make adding new channels easy via a switch case.

**Architecture:** `Notifier` interface with a single `Send` method. `Client` wraps one `Notifier` selected by config type via switch. `Message` struct lives in the notification package with just `Text` and `Image` fields. Callers format their own messages before sending.

**Tech Stack:** Go, slog for logging, existing `network` package for HTTP.

---

### Task 1: Create new Message type and simplify Notifier interface

**Files:**
- Create: `internal/notification/message.go`
- Modify: `internal/notification/interface.go`

- [ ] **Step 1: Create `internal/notification/message.go`**

```go
package notification

// Message represents a notification message ready to send.
type Message struct {
	Text  string // Final formatted text
	Image []byte // Optional image data (nil if no image)
}
```

- [ ] **Step 2: Rewrite `internal/notification/interface.go`**

Replace entire file content with:

```go
package notification

import "context"

// Notifier defines the interface for a notification channel.
type Notifier interface {
	Send(ctx context.Context, message *Message) error
}
```

- [ ] **Step 3: Verify it compiles (expect errors from unupdated files)**

Run: `go vet ./internal/notification/...`
Expected: Errors about `TelegramNotifier` and `Client` not satisfying interfaces — this is correct, we fix them in the next tasks.

- [ ] **Step 4: Commit**

```bash
git add internal/notification/message.go internal/notification/interface.go
git commit -m "refactor(notification): simplify Notifier interface and add Message type"
```

---

### Task 2: Rewrite TelegramNotifier

**Files:**
- Modify: `internal/notification/telegram.go`

- [ ] **Step 1: Rewrite `internal/notification/telegram.go`**

Replace entire file content with:

```go
package notification

import (
	"context"
	"fmt"
	"log/slog"

	"goto-bangumi/internal/model"
	"goto-bangumi/internal/network"
)

// TelegramNotifier sends notifications via Telegram Bot API.
type TelegramNotifier struct {
	token   string
	chatID  string
	baseURL string
}

// NewTelegramNotifier creates a ready-to-use Telegram notifier.
func NewTelegramNotifier(config *model.NotificationConfig) (*TelegramNotifier, error) {
	if config.Token == "" || config.ChatID == "" {
		return nil, fmt.Errorf("telegram requires token and chat_id")
	}
	return &TelegramNotifier{
		token:   config.Token,
		chatID:  config.ChatID,
		baseURL: fmt.Sprintf("https://api.telegram.org/bot%s/", config.Token),
	}, nil
}

// Send sends a notification via Telegram.
func (t *TelegramNotifier) Send(ctx context.Context, message *Message) error {
	if message == nil {
		return fmt.Errorf("message cannot be nil")
	}

	var err error
	if len(message.Image) > 0 {
		err = t.sendPhoto(ctx, message.Text, message.Image)
	} else {
		err = t.sendText(ctx, message.Text)
	}

	if err != nil {
		slog.Error("[Telegram] Failed to send notification", "error", err)
		return err
	}

	return nil
}

func (t *TelegramNotifier) sendPhoto(ctx context.Context, text string, photo []byte) error {
	url := fmt.Sprintf("%ssendPhoto", t.baseURL)

	formData := map[string]string{
		"chat_id":              t.chatID,
		"caption":              text,
		"disable_notification": "true",
	}

	files := map[string][]byte{
		"photo": photo,
	}

	client := network.GetRequestClient()
	_, err := client.PostData(ctx, url, formData, files)
	if err != nil {
		return fmt.Errorf("failed to send photo message: %w", err)
	}

	return nil
}

func (t *TelegramNotifier) sendText(ctx context.Context, text string) error {
	url := fmt.Sprintf("%ssendMessage", t.baseURL)

	formData := map[string]string{
		"chat_id":              t.chatID,
		"text":                 text,
		"disable_notification": "true",
	}

	client := network.GetRequestClient()
	_, err := client.PostData(ctx, url, formData, nil)
	if err != nil {
		return fmt.Errorf("failed to send text message: %w", err)
	}

	return nil
}
```

- [ ] **Step 2: Verify telegram.go compiles**

Run: `go vet ./internal/notification/...`
Expected: Errors from `notification_client.go` still referencing old types — that's expected, fixed in next task.

- [ ] **Step 3: Commit**

```bash
git add internal/notification/telegram.go
git commit -m "refactor(notification): rewrite TelegramNotifier with simplified interface"
```

---

### Task 3: Rewrite Client (fix Send bug, remove processMsg, check Enable)

**Files:**
- Modify: `internal/notification/notification_client.go`

- [ ] **Step 1: Rewrite `internal/notification/notification_client.go`**

Replace entire file content with:

```go
package notification

import (
	"context"
	"log/slog"

	"goto-bangumi/internal/model"
)

// Client wraps a single Notifier selected by configuration.
type Client struct {
	notifier Notifier
}

// NotificationClient is the global notification client.
var NotificationClient = &Client{}

// Init initializes the notification client with the configured channel.
func (c *Client) Init(config *model.NotificationConfig) {
	if !config.Enable {
		slog.Info("[Notification] Notification disabled")
		return
	}

	switch config.Type {
	case "telegram":
		notifier, err := NewTelegramNotifier(config)
		if err != nil {
			slog.Error("[Notification] Failed to init Telegram", "error", err)
			return
		}
		c.notifier = notifier
	default:
		slog.Warn("[Notification] Unknown notification type", "type", config.Type)
	}
}

// Send sends a notification message. Errors are logged but not returned.
func (c *Client) Send(ctx context.Context, message *Message) {
	if c.notifier == nil {
		slog.Warn("[Notification] No notifier initialized, skipping")
		return
	}

	if err := c.notifier.Send(ctx, message); err != nil {
		slog.Error("[Notification] Send failed", "error", err)
	}
}
```

- [ ] **Step 2: Verify notification package compiles**

Run: `go vet ./internal/notification/...`
Expected: PASS (notification package is now self-consistent)

- [ ] **Step 3: Commit**

```bash
git add internal/notification/notification_client.go
git commit -m "fix(notification): actually call notifier.Send, check Enable config, remove processMsg"
```

---

### Task 4: Update caller in rename.go

**Files:**
- Modify: `internal/rename/rename.go:82-91`

- [ ] **Step 1: Update the notification call in `internal/rename/rename.go`**

Replace lines 82-91 (the notification block inside the for loop in `Rename` method):

Old:
```go
		// 发送改名成功通知
		Nclient := notification.NotificationClient
		msg := &model.Message{
			Title:      bangumi.OfficialTitle,
			Season:     strconv.Itoa(bangumi.Season),
			Episode:    strconv.Itoa(metaInfo.Episode),
			PosterLink: bangumi.PosterLink,
		}
		Nclient.Send(ctx, msg)
```

New:
```go
		// 发送改名成功通知
		text := fmt.Sprintf("番剧名称：%s\n季度：第%d季\n更新集数：第%d集",
			bangumi.OfficialTitle, bangumi.Season, metaInfo.Episode)

		var image []byte
		if bangumi.PosterLink != "" {
			image, err = network.LoadImage(ctx, bangumi.PosterLink)
			if err != nil {
				slog.Error("[rename] Failed to download poster", "error", err)
			}
		}

		notification.NotificationClient.Send(ctx, &notification.Message{
			Text:  text,
			Image: image,
		})
```

- [ ] **Step 2: Update imports in `internal/rename/rename.go`**

Remove `"goto-bangumi/internal/model"` import if no longer used elsewhere in the file. Add `"fmt"` and `"goto-bangumi/internal/network"` if not already present. Remove `"strconv"` if no longer used.

Check what else uses `model` in this file: `model.Torrent`, `model.Bangumi`, `model.BangumiRenameConfig` — so `model` stays. Remove `strconv` (no longer used after this change). Add `fmt` and `network`.

Updated imports:
```go
import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"goto-bangumi/internal/database"
	"goto-bangumi/internal/download"
	"goto-bangumi/internal/model"
	"goto-bangumi/internal/network"
	"goto-bangumi/internal/notification"
	"goto-bangumi/internal/parser"
)
```

- [ ] **Step 3: Verify the full project compiles**

Run: `go vet ./...`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/rename/rename.go
git commit -m "refactor(rename): format notification message at call site"
```

---

### Task 5: Delete dead files

**Files:**
- Delete: `internal/notification/plugin/bark.go`
- Delete: `internal/model/notification.go`

- [ ] **Step 1: Delete `internal/notification/plugin/bark.go`**

```bash
rm internal/notification/plugin/bark.go
rmdir internal/notification/plugin
```

- [ ] **Step 2: Delete `internal/model/notification.go`**

```bash
rm internal/model/notification.go
```

- [ ] **Step 3: Verify no remaining references to deleted types**

Run: `go vet ./...`
Expected: PASS (no code references `model.Message` or `model.NewMessage` anymore)

- [ ] **Step 4: Commit**

```bash
git add -A internal/notification/plugin internal/model/notification.go
git commit -m "chore: delete empty bark plugin and unused model.Message"
```

---

### Task 6: Verify everything works end-to-end

- [ ] **Step 1: Run full project build**

Run: `go build ./...`
Expected: PASS

- [ ] **Step 2: Run all tests**

Run: `go test ./...`
Expected: PASS (no notification tests exist currently, but other tests should not break)

- [ ] **Step 3: Final commit if any fixes were needed**

Only commit if previous steps required fixes. Otherwise skip.
