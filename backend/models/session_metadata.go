package models

type SessionMetadata struct {
	ID        uint                   `bson:"_id" json:"id"`                // Session.ID из PostgreSQL
	SessionID uint                   `bson:"session_id" json:"session_id"` // дублируется для удобства
	Type      string                 `bson:"type" json:"type"`             // "cinema", "game", etc.
	Fields    map[string]interface{} `bson:"fields" json:"fields"`         // произвольные поля
	Notes     string                 `bson:"notes,omitempty" json:"notes,omitempty"`
}
