package firebase

import (
	"context"
	"log"
	"math/rand"
	"os"
	"time"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
)

var messagingClient *messaging.Client

func InitFirebase() error {
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		credPath := os.Getenv("FIREBASE_CREDENTIALS_PATH")
		if credPath == "" {
			credPath = "./firebase-credentials.json"
		}
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credPath)
	}

	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		return err
	}

	messagingClient, err = app.Messaging(context.Background())
	if err != nil {
		return err
	}

	log.Println("Firebase initialized successfully using Application Default Credentials")
	return nil
}

func SendPushNotification(token, title, body, imageURL string, data map[string]string) error {
	if messagingClient == nil {
		return nil // Firebase не инициализирован, пропускаем
	}

	message := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title:    title,
			Body:     body,
			ImageURL: imageURL,
		},
		Data: data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				Sound:       "default",
				Priority:    messaging.PriorityHigh,
				ChannelID:   "event_reminders",
				ClickAction: "FLUTTER_NOTIFICATION_CLICK",
			},
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound:            "default",
					Badge:            nil,
					ContentAvailable: true,
					MutableContent:   true,
				},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := messagingClient.Send(ctx, message)
	if err != nil {
		return err
	}

	log.Printf("Successfully sent FCM notification: %s", response)
	return nil
}

func GetRandomEventTitle() string {
	titles := []string{
		"Делу время, а событие через...",
		"Без труда и на событие не попадешь!",
		"Время не ждёт — событие на подходе!",
		"Скоро начнётся что-то интересное...",
		"Не опаздывай! Событие уже близко",
		"Твоё приключение скоро начнётся!",
		"Готовься! Событие не за горами",
		"Часики тикают — событие приближается",
		"Пора собираться! Событие скоро",
		"Напоминаем: событие вот-вот начнётся",
		"Время пришло! Событие уже близко",
		"Не пропусти! Событие совсем скоро",
		"Событие на горизонте — готовься!",
		"Твоё время пришло! Событие близко",
		"Скоро, очень скоро начнётся...",
		"Событие стучится в дверь!",
		"Приготовься к событию дня!",
		"Событие уже на пороге!",
		"Время событий настало!",
		"Не забудь! Событие уже рядом",
		"Событие не заставит себя ждать!",
		"Готовься к яркому событию!",
		"Событие вот-вот стартует!",
		"Осталось совсем чуть-чуть...",
		"Событие дня уже близко!",
		"Пора в путь! Событие скоро",
		"Не проспи! Событие приближается",
		"Твоё событие почти здесь!",
		"Время события пришло!",
		"Событие уже на старте!",
		"Скоро начнётся нечто особенное!",
		"Будь готов! Событие близко",
		"Событие ждёт тебя!",
		"Приключение вот-вот начнётся!",
		"Событие на подходе — не опоздай!",
		"Готовься! Твоё время пришло",
		"Событие дня стартует скоро!",
		"Не забывай про событие!",
		"Скоро всё начнётся!",
		"Событие уже совсем близко!",
		"Время действовать! Событие скоро",
	}

	rand.NewSource(time.Now().UnixNano())
	return titles[rand.Intn(len(titles))]
}
