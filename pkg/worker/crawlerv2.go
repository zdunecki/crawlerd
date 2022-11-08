package worker

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	urllib "net/url"
)

// TODO: robots.txt, seed, sitemap (sitemap.xml, sitemap_index.xml, feed.atom)

var siteMapRoots = map[string]bool{
	"sitemap.xml":       true,
	"sitemap_index.xml": true,
	"feed.atom":         true,
}

type CrawlerV2Error struct {
	err error
}

func (err *CrawlerV2Error) Error() string {
	return err.err.Error()
}

var ErrCrawlerV2InvalidHTTPStatus = errors.New("invalid http status")

type CrawlerV2CrawlSpec struct {
	Seeds     []string
	Extracter ExtracterAPI
}

type CrawlerV2 struct {
	http *http.Client
	log  *log.Entry
}

func NewCrawlerV2() *CrawlerV2 {
	return &CrawlerV2{}
}

func (c *CrawlerV2) Crawl(spec *CrawlerV2CrawlSpec) {
	fetcher := func() {

	}

	for _, seed := range spec.Seeds {
		resp, err := c.fetch(seed)
		if err != nil {
			if resp != nil {
				c.log.Error(resp)
			}

			c.log.Error(err)

			continue
		}

		if err := c.parseBody(spec.Extracter, resp.Body); err != nil {
			c.log.Error(err)

			continue
		}

		for root, ok := range siteMapRoots {
			if !ok {
				continue
			}

			sitemapRoot, err := urllib.JoinPath(seed, root)
			if err != nil {
				c.log.Error(err)
				continue
			}
			
			fmt.Println(sitemapRoot)

		}

	}
}

// TODO: pipe from goquery to extracter
func (c *CrawlerV2) parseBody(extracter ExtracterAPI, body io.Reader) error {
	if extracter == nil {
		return nil
	}

	return extracter.Extract(body) // TODO: for client-side extract get stream of new content
}

func (c *CrawlerV2) fetch(url string) (*http.Response, error) {
	_, err := urllib.Parse(url)
	if err != nil {
		return nil, err
	}

	head, err := c.http.Head(url)
	if err != nil {
		return nil, err
	}

	// TODO: handle 3xx

	if head.StatusCode >= http.StatusBadRequest {
		return head, &CrawlerV2Error{
			err: ErrCrawlerV2InvalidHTTPStatus,
		}
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	return c.http.Do(req)
}
