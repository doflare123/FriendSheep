package db

import (
	"context"
	"fmt"
	"friendship/models/sessions"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetSessionsMetadata(sessionIDs []uint) (map[uint]*sessions.SessionMetadata, error) {
	// Получаем MongoDB базу данных (не коллекцию!)
	mongoDatabase := GetMongoDB()
	if mongoDatabase == nil {
		return make(map[uint]*sessions.SessionMetadata), nil
	}

	// ИСПРАВЛЕНИЕ: Получаем коллекцию из базы данных
	collection := mongoDatabase.Collection("session_metadata")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Преобразуем []uint в []interface{} для MongoDB запроса
	sessionIDsInterface := make([]interface{}, len(sessionIDs))
	for i, id := range sessionIDs {
		sessionIDsInterface[i] = id
	}

	filter := bson.M{
		"session_id": bson.M{"$in": sessionIDsInterface},
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return make(map[uint]*sessions.SessionMetadata), nil
		}
		return nil, err
	}
	defer cursor.Close(ctx)

	metadataMap := make(map[uint]*sessions.SessionMetadata)
	for cursor.Next(ctx) {
		var metadata sessions.SessionMetadata
		if err := cursor.Decode(&metadata); err != nil {
			continue // пропускаем ошибочные документы
		}
		metadataMap[metadata.SessionID] = &metadata
	}

	return metadataMap, cursor.Err()
}

func GetSessionMetadataId(sessionID *uint) (*sessions.SessionMetadata, error) {
	if sessionID == nil {
		return nil, fmt.Errorf("sessionID не может быть nil")
	}

	mongoDatabase := GetMongoDB()
	if mongoDatabase == nil {
		return nil, nil // возвращаем nil без ошибки, если MongoDB недоступна
	}

	collection := mongoDatabase.Collection("session_metadata")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"session_id": *sessionID}

	var metadata sessions.SessionMetadata
	err := collection.FindOne(ctx, filter).Decode(&metadata)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // метаданные не найдены - это нормально
		}
		return nil, err
	}

	return &metadata, nil
}
