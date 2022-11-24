package crawler

import (
	"github.com/zdunecki/crawlerd/pkg/extract"
	"testing"
)

func TestCrawlerV2(t *testing.T) {
	crawler := New()

	{
		spec := &Spec[*extract.SitemapResponse]{
			Seeds:     []string{"http://livesession.io/sitemap.xml"},
			Extracter: extract.NewSitemap(),
		}

		respC, err := crawler.CrawlBlogSitemap(spec)
		if err != nil {
			t.Fatal(err)
		}

		for resp := range respC {
			if resp.Error != nil {
				t.Error(resp.Error)
			} else {
				t.Log(resp.Links)
			}
		}
	}

	{
		spec := &Spec[*extract.ArticleResponse]{
			Seeds:     []string{"http://livesession.io"},
			Extracter: extract.NewArticle(&extract.ArticleSpec{}),
		}

		respC, err := crawler.CrawlBlog(spec)
		if err != nil {
			t.Fatal(err)
		}

		for resp := range respC {
			if resp.Error != nil {
				t.Error(resp.Error)
			}
		}
	}
}
