package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Bot представляет Telegram бота
type Bot struct {
	token  string
	apiURL string
	client *http.Client
}

// NotificationMessage структура сообщения уведомления
type NotificationMessage struct {
	Title    string `json:"title"`
	Text     string `json:"text"`
	ImageURL string `json:"image_url"`
}

// TelegramMessage структура для отправки сообщения в Telegram API
type TelegramMessage struct {
	ChatID      string      `json:"chat_id"`
	Text        string      `json:"text"`
	ParseMode   string      `json:"parse_mode,omitempty"`
	Photo       string      `json:"photo,omitempty"`
	Caption     string      `json:"caption,omitempty"`
	ReplyMarkup interface{} `json:"reply_markup,omitempty"`
}

// TelegramPhotoMessage структура для отправки фото
type TelegramPhotoMessage struct {
	ChatID      string      `json:"chat_id"`
	Photo       string      `json:"photo"`
	Caption     string      `json:"caption"`
	ParseMode   string      `json:"parse_mode,omitempty"`
	ReplyMarkup interface{} `json:"reply_markup,omitempty"`
}

// NewBot создает новый экземпляр бота
func NewBot() *Bot {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		panic("TELEGRAM_BOT_TOKEN environment variable is required")
	}

	return &Bot{
		token:  token,
		apiURL: fmt.Sprintf("https://api.telegram.org/bot%s", token),
		client: &http.Client{},
	}
}

// SendNotification отправляет уведомление пользователю
func (b *Bot) SendNotification(telegramID string, message NotificationMessage) error {
	// Если есть картинка, отправляем как фото с подписью
	if message.ImageURL != "" {
		return b.sendPhotoMessage(telegramID, message)
	}

	// Иначе отправляем обычное текстовое сообщение
	return b.sendTextMessage(telegramID, message)
}

// sendPhotoMessage отправляет сообщение с фотографией
func (b *Bot) sendPhotoMessage(telegramID string, message NotificationMessage) error {
	caption := fmt.Sprintf("*%s*\n\n%s", message.Title, message.Text)

	photoMessage := TelegramPhotoMessage{
		ChatID:    telegramID,
		Photo:     message.ImageURL,
		Caption:   caption,
		ParseMode: "Markdown",
	}

	return b.sendRequest("sendPhoto", photoMessage)
}

// sendTextMessage отправляет текстовое сообщение
func (b *Bot) sendTextMessage(telegramID string, message NotificationMessage) error {
	text := fmt.Sprintf("*%s*\n\n%s", message.Title, message.Text)

	textMessage := TelegramMessage{
		ChatID:    telegramID,
		Text:      text,
		ParseMode: "Markdown",
	}

	return b.sendRequest("sendMessage", textMessage)
}

// sendRequest выполняет запрос к Telegram API
func (b *Bot) sendRequest(method string, payload interface{}) error {
	url := fmt.Sprintf("%s/%s", b.apiURL, method)

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := b.client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API error: status %d, response: %s", resp.StatusCode, string(body))
	}

	// Проверяем успешность ответа
	var response struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("error parsing response: %w", err)
	}

	if !response.OK {
		return fmt.Errorf("telegram API returned error: %s", response.Description)
	}

	return nil
}

// GetWebhookInfo получает информацию о webhook (для проверки настроек бота)
func (b *Bot) GetWebhookInfo() (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/getWebhookInfo", b.apiURL)

	resp, err := b.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error getting webhook info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return result, nil
}

// GetMe получает информацию о боте
func (b *Bot) GetMe() (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/getMe", b.apiURL)

	resp, err := b.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error getting bot info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return result, nil
}
