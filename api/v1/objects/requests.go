package objects

// TODO: use struct from meta pkg
type RequestPostURL struct {
	URL string `json:"url"`
	// Deprecated: Interval is not important now
	Interval int `json:"interval"`
}

type RequestPatchURL struct {
	URL *string `json:"url" bson:"url,omitempty"`

	// Deprecated: Interval is not important now
	Interval *int `json:"interval" bson:"interval,omitempty"`
}
