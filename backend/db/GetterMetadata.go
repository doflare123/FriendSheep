package db

import (
	"context"
	"fmt"
	"friendship/models/events"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetSessionsMetadata(sessionIDs []uint) (map[uint]*events.SessionMetadata, error) {
	// Получаем MongoDB базу данных (не коллекцию!)
	mongoDatabase := GetMongoDB()
	if mongoDatabase == nil {
		return make(map[uint]*events.SessionMetadata), nil
	}

	collection := mongoDatabase.Collection("session_metadata")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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
			return make(map[uint]*events.SessionMetadata), nil
		}
		return nil, err
	}
	defer cursor.Close(ctx)

	metadataMap := make(map[uint]*events.SessionMetadata)
	for cursor.Next(ctx) {
		var metadata events.SessionMetadata
		if err := cursor.Decode(&metadata); err != nil {
			continue // пропускаем ошибочные документы
		}
		metadataMap[metadata.SessionID] = &metadata
	}

	return metadataMap, cursor.Err()
}

func GetSessionMetadataId(sessionID *uint) (*events.SessionMetadata, error) {
	if sessionID == nil {
		return nil, fmt.Errorf("sessionID не может быть nil")
	}

	mongoDatabase := GetMongoDB()
	if mongoDatabase == nil {
		return nil, nil
	}

	collection := mongoDatabase.Collection("session_metadata")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"session_id": *sessionID}

	var metadata events.SessionMetadata
	err := collection.FindOne(ctx, filter).Decode(&metadata)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &metadata, nil
}
