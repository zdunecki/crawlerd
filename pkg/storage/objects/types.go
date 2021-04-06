package objects

type (
	URL struct {
		ID       int    `json:"id" bson:"id"`
		URL      string `json:"url" bson:"url"`
		Interval int    `json:"interval" bson:"interval"`
	}
	History struct {
		Response        string  `json:"response" bson:"response"`
		CreatedAt       float64 `json:"created_at" bson:"created_at"`
		DurationSeconds float64 `json:"duration" bson:"duration"`
	}
	Sequence struct {
		ObjectID string `bson:"_id"`
		ID       int    `bson:"id"`
	}
)
