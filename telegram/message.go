package telegram

// TelegramMessage represents the structure of a message to be sent to Telegram.
type TelegramMessage struct {
	ChatID          string `json:"chat_id"`
	Text            string `json:"text"`
	MessageThreadID string `json:"message_thread_id,omitempty"`
}
