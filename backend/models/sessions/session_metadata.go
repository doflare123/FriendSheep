package sessions

//mongodb

type SessionMetadata struct {
	SessionID uint                   `bson:"session_id"`
	Type      string                 `bson:"type"`
	Fields    map[string]interface{} `bson:"fields"`
	Location  string                 `bson:"location"`
	Genres    []string               `bson:"genres,omitempty"`
	Year      *int                   `bson:"year,omitempty"`
	Country   string                 `bson:"country,omitempty"`
	AgeLimit  string                 `bson:"age_limit,omitempty"`
	Notes     string                 `bson:"notes,omitempty"`
}
