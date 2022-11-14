package extract

import (
	"net/http"
	"os"
	"testing"
)

func TestExtractSitemap(t *testing.T) {
	extractor := NewSitemap()

	resp, err := http.Get("https://livesession.io/sitemap.xml")
	if err != nil {
		t.Fatal(err)
	}

	data, err := extractor.Extract(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if len(data.Links) <= 0 {
		t.Fatal("sitemap should have links")
	}
}

func TestExtractArticle(t *testing.T) {
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "../../gcp.json")
	extractor := NewArticle(nil, WithArticleWithGCPClassifier())

	u := "https://blog.allegro.tech/2021/01/impact-of-the-data-model-on-the-MongoDB-database-size.html"
	resp, err := http.Get(u)
	if err != nil {
		t.Fatal(err)
	}

	data, err := extractor.Extract(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if data.CategoryRoot != ArticleCategoryEngineering {
		t.Fatal("category root should be engineering")
	}
}
