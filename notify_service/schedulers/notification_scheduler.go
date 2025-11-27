package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"gorm.io/gorm"

	"notify_service/models"
	"notify_service/models/sessions"
)

type TelegramMessage struct {
	Items []TelegramItem `json:"items"`
}

type TelegramItem struct {
	TelegramIDs []int64 `json:"telegramIds"`
	ImageURL    string  `json:"imageUrl"`
	Title       string  `json:"title"`
	Text        string  `json:"text"`
}

func StartNotificationWorker(db *gorm.DB) {
	ticker := time.NewTicker(1 * time.Minute)
	fmt.Print("StartNotificationWorker")
	go func() {
		for range ticker.C {
			processSessions(db)
		}
	}()
}

func processSessions(db *gorm.DB) {
	now := time.Now()
	log.Println("=== processSessions started at", now)

	statusIDs := map[string]uint{}
	var statuses []sessions.Status
	if err := db.Find(&statuses).Error; err != nil {
		log.Println("Error fetching statuses:", err)
		return
	}
	for _, st := range statuses {
		statusIDs[st.Status] = st.ID
	}
	log.Println("Loaded statuses:", statusIDs)

	var types []sessions.NotificationType
	if err := db.Find(&types).Error; err != nil {
		log.Println("Error fetching notification types:", err)
		return
	}
	log.Printf("Loaded %d notification types\n", len(types))

	var sessionsList []sessions.Session
	if err := db.Preload("User").
		Where("status_id = ?", statusIDs["Набор"]).
		Find(&sessionsList).Error; err != nil {
		log.Println("Error fetching sessions:", err)
		return
	}
	log.Printf("Found %d sessions with status 'Набор'\n", len(sessionsList))

	for _, s := range sessionsList {
		for _, nt := range types {
			notifyTime := s.StartTime.Add(-time.Duration(nt.HoursBefore) * time.Hour)

			if now.After(notifyTime.Add(-30*time.Second)) && now.Before(notifyTime.Add(30*time.Second)) {
				var exists int64
				if err := db.Model(&sessions.Notification{}).
					Where("session_id = ? AND notification_type_id = ?", s.ID, nt.ID).
					Count(&exists).Error; err != nil {
					log.Println("Error checking notification existence:", err)
					continue
				}
				if exists > 0 {
					log.Printf("Notification for session %d and type %d already sent\n", s.ID, nt.ID)
					continue
				}

				log.Printf("Sending notification: session=%d, type=%s, notifyTime=%v\n", s.ID, nt.Name, notifyTime)
				sendSessionNotification(db, s, nt)
			}
		}
	}

	if err := db.Model(&sessions.Session{}).
		Where("status_id = ?", statusIDs["Набор"]).
		Where("start_time <= ?", now).
		Update("status_id", statusIDs["В процессе"]).Error; err != nil {
		log.Println("Error updating sessions to 'В процессе':", err)
	}

	if err := db.Model(&sessions.Session{}).
		Where("status_id = ?", statusIDs["В процессе"]).
		Where("end_time <= ?", now).
		Update("status_id", statusIDs["Завершена"]).Error; err != nil {
		log.Println("Error updating sessions to 'В процессе':", err)
	}

	var completedSessions []sessions.Session
	if err := db.Where("status_id = ?", statusIDs["Завершена"]).
		Where("end_time <= ?", now).
		Find(&completedSessions).Error; err == nil {

		for _, cs := range completedSessions {
			notifyStatisticsUpdate(cs.ID)
		}
	}
}

func sendSessionNotification(db *gorm.DB, s sessions.Session, nt sessions.NotificationType) {
	var participants []sessions.SessionUser
	if err := db.Where("session_id = ?", s.ID).Find(&participants).Error; err != nil {
		log.Println("Error fetching participants:", err)
		return
	}

	telegramIDs := []int64{}
	for _, p := range participants {
		var user models.User
		if err := db.First(&user, p.UserID).Error; err == nil && user.TelegramID != nil {
			if tid, err := parseTelegramID(*user.TelegramID); err == nil {
				telegramIDs = append(telegramIDs, tid)
			}
		}
	}

	if len(telegramIDs) == 0 {
		log.Printf("No Telegram IDs found for session %d\n", s.ID)
		return
	}

	notifyTypeText := map[string]string{
		"24_hours": "24 часа",
		"6_hours":  "6 часов",
		"1_hour":   "1 час",
	}

	text := fmt.Sprintf(
		"Напоминаем, что мероприятие \"%s\" начнется через %s",
		s.Title,
		notifyTypeText[nt.Name],
	)

	notif := sessions.Notification{
		UserID:             s.UserID,
		SessionID:          s.ID,
		NotificationTypeID: nt.ID,
		SendAt:             time.Now(),
		Sent:               true,
		Title:              "Напоминание о мероприятии",
		Text:               text,
		ImageURL:           s.ImageURL,
	}
	if err := db.Create(&notif).Error; err != nil {
		log.Println("Error creating notification:", err)
	}

	msg := TelegramMessage{
		Items: []TelegramItem{
			{
				TelegramIDs: telegramIDs,
				ImageURL:    s.ImageURL,
				Title:       "Напоминание о мероприятии",
				Text:        text,
			},
		},
	}

	sendToTelegramBot(msg)
}

func sendToTelegramBot(msg TelegramMessage) {
	url := os.Getenv("URL_BOT")
	apiKey := os.Getenv("TELEGRAM_BOT_API_KEY")

	data, _ := json.Marshal(msg)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		log.Println("Error creating request to telegram bot:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending to telegram bot:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Telegram bot returned status %d\n", resp.StatusCode)
	}
}

func parseTelegramID(idStr string) (int64, error) {
	return strconv.ParseInt(idStr, 10, 64)
}

func notifyStatisticsUpdate(sessionID uint) {
	url := os.Getenv("FRIENDSHIP_URL") + "/internal/update-statistics"
	token := os.Getenv("NOTIFY_SERVICE_TOKEN")

	payload := map[string]uint{"session_id": sessionID}
	data, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		log.Println("Error creating request to friendship service:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Token", token)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending statistics update:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Statistics update returned status %d\n", resp.StatusCode)
	}
}
