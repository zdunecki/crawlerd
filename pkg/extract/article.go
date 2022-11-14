package extract

import (
	"bytes"
	language "cloud.google.com/go/language/apiv1"
	"cloud.google.com/go/language/apiv1/languagepb"
	"context"
	goose "github.com/advancedlogic/GoOse"
	log "github.com/sirupsen/logrus"
	"io"
	"strings"
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

type articleOptions struct {
	classifier func([]byte) ([]*ArticleCategory, error)
}

const (
	ArticleCategoryEngineering = "engineering"
)

func classifierGCP(client *language.Client) func(content []byte) ([]*ArticleCategory, error) {
	return func(content []byte) ([]*ArticleCategory, error) {
		classify, err := client.ClassifyText(context.Background(), &languagepb.ClassifyTextRequest{
			Document: &languagepb.Document{
				Source: &languagepb.Document_Content{
					Content: string(content),
				},
				Type: languagepb.Document_PLAIN_TEXT,
			},
		})

		if err != nil {
			return nil, err
		}

		categories := make([]*ArticleCategory, 0)

		if classify.Categories != nil && len(classify.Categories) > 0 {
			for _, category := range classify.Categories {
				id := ""
				name := ""
				var score float32 = 0.0

				if strings.Contains(category.Name, "Computers & Electronics") {
					id = ArticleCategoryEngineering
					name = "Engineering"
					score = category.GetConfidence()
				}

				if id != "" {
					categories = append(categories, &ArticleCategory{
						ID:    id,
						Name:  name,
						Score: score,
					})
				}
			}
		}

		return categories, nil
	}
}

func WithArticleWithGCPClassifier() func(*articleOptions) func() {
	ctx := context.Background()

	client, err := language.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	return func(options *articleOptions) func() {
		options.classifier = classifierGCP(client)

		return func() {
			client.Close()
		}
	}
}

// TODO: options like regexp etc.
func NewArticle(spec *ArticleSpec, options ...func(options *articleOptions) func()) API[*ArticleResponse] {
	opt := &articleOptions{}

	doneCallbacks := make([]func(), 0)
	done := func() {
		for _, f := range doneCallbacks {
			f()
		}
	}

	for _, o := range options {
		doneCallbacks = append(doneCallbacks, o(opt))
	}

	return &Article{
		Spec:    spec,
		options: opt,
		done:    done,
	}
}

type Article struct {
	Spec    *ArticleSpec
	options *articleOptions
	done    func()
}

// TODO: cache for crawler and classifier
func (eb *Article) Extract(reader io.Reader) (*ArticleResponse, error) {
	crawler := goose.NewCrawler(goose.GetDefaultConfiguration())

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	article, err := crawler.Crawl(string(data), "")
	if err != nil {
		return nil, err
	}

	categoryRoot := ""
	articleCategories := make([]*ArticleCategory, 0)

	if eb.options.classifier != nil {
		categories, err := eb.options.classifier([]byte(article.CleanedText))
		if err != nil {
			return nil, err
		}

		articleCategories = categories

		if len(categories) > 0 {
			categoryRoot = categories[0].ID
		}
	}

	content := io.NopCloser(bytes.NewReader([]byte(article.CleanedText)))

	return &ArticleResponse{
		Type:         "",
		Title:        article.Title,
		URL:          article.CanonicalLink,
		Icon:         article.MetaFavicon,
		Sitename:     article.MetaDescription,
		Date:         article.PublishDate,
		Author:       "",
		Sentiment:    0,
		Images:       nil,
		CategoryRoot: categoryRoot,
		Categories:   articleCategories,
		Content:      content,
		Page:         io.NopCloser(bytes.NewReader([]byte(article.RawHTML))),
		Tags:         nil,
		Error:        nil,
	}, nil
}
