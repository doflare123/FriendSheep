package events

//mongodb

type SessionMetadata struct {
	SessionID uint                   `bson:"session_id"`
	Fields    map[string]interface{} `bson:"fields"`
	Location  string                 `bson:"location"`
	Genres    []string               `bson:"genres,omitempty"`
	Year      *int                   `bson:"year,omitempty"`
	Country   string                 `bson:"country,omitempty"`
	AgeLimit  string                 `bson:"age_limit,omitempty"`
	Notes     string                 `bson:"notes,omitempty"`
}
