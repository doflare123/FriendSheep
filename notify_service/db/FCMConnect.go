package db

import (
	"context"
	"errors"
	"fmt"
	"os"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"google.golang.org/api/option"
)

var fcmClient *messaging.Client

func InitFCM() error {
	ctx := context.Background()

	serviceAccountKey := os.Getenv("FIREBASE_SERVICE_ACCOUNT_KEY")
	if serviceAccountKey == "" {
		return errors.New("FIREBASE_SERVICE_ACCOUNT_KEY environment variable not set")
	}

	opt := option.WithCredentialsFile(serviceAccountKey)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return err
	}
	fcmClient, err = app.Messaging(ctx)
	if err != nil {
		return err
	}

	return nil
}

func GetFCM() *messaging.Client {
	return fcmClient
}

func SendNotification(ctx context.Context, token string, title, body string, data map[string]string) (*messaging.SendResponse, error) {
	if fcmClient == nil {
		return nil, fmt.Errorf("FCM client is not initialized")
	}

	message := &messaging.Message{
		Data: data,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Token: token,
	}

	response, err := fcmClient.Send(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %v", err)
	}

	return &messaging.SendResponse{
		MessageID: response,
	}, nil
}

// отправляет уведомление нескольким получателям
func SendMulticastNotification(ctx context.Context, tokens []string, title, body string, data map[string]string) (*messaging.BatchResponse, error) {
	if fcmClient == nil {
		return nil, fmt.Errorf("FCM client is not initialized")
	}

	message := &messaging.MulticastMessage{
		Data: data,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Tokens: tokens,
	}

	response, err := fcmClient.SendMulticast(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("failed to send multicast message: %v", err)
	}

	return response, nil
}
