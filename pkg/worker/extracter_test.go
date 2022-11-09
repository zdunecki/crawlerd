package worker

import (
	"net/http"
	"testing"
)

func TestExtractor(t *testing.T) {
	extractor := &ExtracterBlogSitemap{}

	resp, err := http.Get("https://fullstory.com/sitemap.xml")
	if err != nil {
		t.Fatal(err)
	}

	extractor.Extract(resp.Body)
}
