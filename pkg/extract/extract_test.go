package extract

import (
	"net/http"
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
