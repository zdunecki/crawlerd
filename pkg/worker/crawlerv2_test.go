package worker

import "testing"

func TestCrawlerV2(t *testing.T) {
	crawler := NewCrawlerV2()

	{
		spec := &CrawlerV2CrawlSpec[*ExtracterBlogSitemapResponse]{
			Seeds:     []string{"http://livesession.io"},
			Extracter: NewExtracterBlogSitemap(),
		}

		respC, err := crawler.CrawlBlogSitemap(spec)
		if err != nil {
			t.Fatal(err)
		}

		for resp := range respC {
			if resp.Error != nil {
				t.Error(resp.Error)
			}
		}
	}

	{
		spec := &CrawlerV2CrawlSpec[*ExtracterBlogResponse]{
			Seeds: []string{"http://livesession.io"},
			Extracter: NewExtracterBlog(&ExtracterBlogSpec{
				IgnoreSiteMap: true,
			}),
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
