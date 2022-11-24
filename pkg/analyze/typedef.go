package analyze

type Rect struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`

	Width  float64 `json:"width"`
	Height float64 `json:"height"`

	Top    float64 `json:"top"`
	Bottom float64 `json:"bottom"`

	Left  float64 `json:"left"`
	Right float64 `json:"right"`
}

type DOM struct {
	ID          int                    `json:"i"`
	Type        int                    `json:"nT"`
	TagName     string                 `json:"tN"`
	Attrs       map[string]interface{} `json:"a"`
	TextContent string                 `json:"tC"`
	ChildNodes  []*DOM                 `json:"cN"`
	Rect        *Rect                  `json:"r"`
}

const (
	JSNodeElement      = 1
	JSNodeAttribute    = 2
	JSNodeText         = 3
	JSNodeCdata        = 4
	JSNodeProcessing   = 7
	JSNodeComment      = 8
	JSNodeDocument     = 9
	JSNodeDocumentType = 10
	JSNodeFragment     = 11
)
