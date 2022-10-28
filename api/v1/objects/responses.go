package objects

type (
	ResponsePostURL struct {
		ID int `json:"id"`
	}
)

// below dont move/delete

type ResponseRequestQueueCreate struct {
	IDs []string `json:"ids"`
}

type ResponseAppendSeed struct {
	OK bool `json:"ok"`
}
