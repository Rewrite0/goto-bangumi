# Notification System Redesign

## Goal

Simplify the notification system architecture, make it easy to add new notification channels, and fix existing bugs. The system supports multiple channels but only sends to one user-configured channel at a time.

## Current Problems

1. `Client.Send()` processes the message but never calls `notifier.PostMsg()` — notifications are not actually sent
2. `processMsg` mixes concerns: message formatting, image downloading, and business logic (episode text assembly) all in one method
3. `Message` model couples business fields (Title/Season/Episode) with transport data (Text/File)
4. `Notifier` interface has `Init` method — initialization should happen at construction time
5. Bark plugin file exists but is empty (and has a typo in the package name: `nofification`)
6. `NotificationConfig.Enable` field is never checked
7. Errors are silently swallowed in `processMsg`

## Design Decisions

- **Single channel**: User configures one notification type; system sends to that one only
- **Caller formats messages**: The notification system only sends — caller is responsible for assembling text and downloading images
- **Keep global singleton**: `NotificationClient` stays as a package-level var, no DI for now
- **Errors logged internally**: `Send` failures are logged but not returned to caller — notification failure should not affect main flow
- **Switch-based dispatch**: `Client.Init` uses a switch to select the notifier implementation; new channels add a case

## Architecture

### Notifier Interface

Simplified to a single method — no `Init`, construction handles initialization:

```go
type Notifier interface {
    Send(ctx context.Context, message *Message) error
}
```

### Message

Moved from `model` package into `notification` package. Only contains what the notification system needs:

```go
type Message struct {
    Text  string // Final formatted text to send
    Image []byte // Optional image data (nil if no image)
}
```

### Client

```go
type Client struct {
    notifier Notifier
}

var NotificationClient = &Client{}

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
        slog.Warn("[Notification] Unknown type", "type", config.Type)
    }
}

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

### Telegram Notifier

Constructor takes config and returns a ready-to-use notifier:

```go
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
```

Internal fields replace the separate `TeleConfig` struct — no need for an extra type when the notifier owns its config.

### Adding a New Channel

To add e.g. Bark:

1. Create `internal/notification/bark.go`
2. Implement `Notifier` interface with `Send(ctx, *Message) error`
3. Add constructor `NewBarkNotifier(config) (*BarkNotifier, error)`
4. Add `case "bark":` in `Client.Init` switch
5. Add any channel-specific config fields to `NotificationConfig` if needed

### Caller-Side Changes

`rename.go` currently passes structured data (Title/Season/Episode/PosterLink) and relies on `processMsg` to format. After this change, the caller assembles the final text and downloads the image itself:

```go
// In rename.go
text := fmt.Sprintf("番剧名称：%s\n季度：第%d季\n更新集数：第%d集",
    bangumi.OfficialTitle, bangumi.Season, metaInfo.Episode)

var image []byte
if bangumi.PosterLink != "" {
    image, err = network.LoadImage(ctx, bangumi.PosterLink)
    if err != nil {
        slog.Error("[rename] Failed to download poster", "error", err)
        // continue without image
    }
}

notification.NotificationClient.Send(ctx, &notification.Message{
    Text:  text,
    Image: image,
})
```

## File Changes Summary

| File | Action |
|------|--------|
| `internal/notification/interface.go` | Simplify: remove `Init`, rename `PostMsg` to `Send`, use `*Message` |
| `internal/notification/notification_client.go` | Remove `processMsg`, fix `Send` to call notifier, check `Enable` |
| `internal/notification/message.go` | New file: `Message` struct (Text + Image) |
| `internal/notification/telegram.go` | Inline config into struct fields, constructor takes `NotificationConfig` |
| `internal/notification/plugin/bark.go` | Delete (empty file with typo) |
| `internal/model/notification.go` | Delete (`Message` moves to notification package) |
| `internal/rename/rename.go` | Format message and download image at call site |

## Out of Scope

- DI injection for NotificationClient
- Multi-channel broadcasting (sending to multiple channels simultaneously)
- Message retry / queue
- Specific new channel implementations (Bark, Webhook, etc.)
