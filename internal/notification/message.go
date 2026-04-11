package notification

// Message represents a notification message ready to send.
type Message struct {
	Text  string // Final formatted text
	Image []byte // Optional image data (nil if no image)
}
