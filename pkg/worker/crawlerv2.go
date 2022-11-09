package worker

import (
	"errors"
	"fmt"
	"github.com/hashicorp/go-multierror"
	log "github.com/sirupsen/logrus"
	"net/http"
	urllib "net/url"
	"time"
)

// TODO: robots.txt, seed, sitemap (sitemap.xml, sitemap_index.xml, feed.atom)

var siteMapRoots = map[string]bool{
	"sitemap.xml":       true,
	"sitemap_index.xml": true,
	"feed.atom":         true,
}

type CrawlerV2HttpError struct {
	*http.Response
}

func (err *CrawlerV2HttpError) Error() string {
	return fmt.Sprintf("status_code:%d", err.StatusCode)
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
	return &CrawlerV2{
		http: &http.Client{
			Timeout: time.Minute,
		},
		log: log.WithField("component", "crawlerv2"),
	}
}

func (c *CrawlerV2) Crawl(spec *CrawlerV2CrawlSpec) {
	for _, seed := range spec.Seeds {
		switch t := spec.Extracter.(type) {
		case *ExtracterBlog:
			if _, err := c.bodyParser(seed, t); err != nil {
				c.log.Error(err)
				continue
			}

			if err := c.siteMapSearchCrawler(t, seed); err != nil {
				c.log.Error(err)
				continue
			}
		}
	}
}

func (c *CrawlerV2) siteMapSearchCrawler(extracter *ExtracterBlog, url string) error {
	var errors error

	for sitemap, ok := range siteMapRoots {
		if !ok {
			continue
		}

		sitemapURL, err := urllib.JoinPath(url, sitemap)
		if err != nil {
			multierror.Append(errors, err)
			continue
		}

		sitemapIndex, err := c.bodyParser(sitemapURL, &ExtracterBlogSitemap{
			Blog: extracter,
		})
		if err != nil {
			multierror.Append(errors, err)
			continue
		}

		switch t := sitemapIndex.(type) {
		case *ExtracterBlogSitemapResponse:
			for _, link := range t.Links {
				if _, err := c.bodyParser(link, extracter); err != nil {
					c.log.Error(err)
					continue
				}
			}
		}
	}

	return errors
}

// TODO: pipe from goquery to extracter
func (c *CrawlerV2) bodyParser(url string, extracter ExtracterAPI) (interface{}, error) {
	resp, err := c.fetch(url)
	if err != nil {
		return nil, err
	}

	if extracter == nil {
		return nil, nil
	}

	return extracter.Extract(resp.Body) // TODO: for client-side extract get stream of new content
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
		return nil, multierror.Append(nil, &CrawlerV2Error{
			err: ErrCrawlerV2InvalidHTTPStatus,
		}, &CrawlerV2HttpError{head})
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	return c.http.Do(req)
}
