package extract

import (
	"io"
	"time"
)

type ArticleImage struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

type ArticleTag struct {
}

type ArticleCategory struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Score float32 `json:"score"`
}

type ArticleSpec struct {
	DeepLink bool // TODO: search deep links
}

type ArticleResponse struct {
	Type         string             `json:"type"`
	Title        string             `json:"title"`
	URL          string             `json:"url"`
	Icon         string             `json:"icon"`
	Sitename     string             `json:"sitename"`
	Date         *time.Time         `json:"date"`
	Author       string             `json:"author"`
	Sentiment    float32            `json:"sentiment"`
	Images       []*ArticleImage    `json:"images"`
	CategoryRoot string             `json:"category_root"`
	Categories   []*ArticleCategory `json:"categories"`
	Content      io.ReadCloser      `json:"content"`
	Page         io.ReadCloser      `json:"page"`
	Tags         []*ArticleTag      `json:"tags"`
	Error        error
}

// TODO: options like regexp etc.
func NewArticle(spec *ArticleSpec) API[*ArticleResponse] {
	return &Article{
		Spec: spec,
	}
}

type Article struct {
	Spec *ArticleSpec
}

func (eb *Article) Extract(reader io.Reader) (*ArticleResponse, error) {
	return &ArticleResponse{}, nil
}
