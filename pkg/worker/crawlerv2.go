package worker

import (
	"errors"
	"fmt"
	"github.com/hashicorp/go-multierror"
	log "github.com/sirupsen/logrus"
	"net/http"
	urllib "net/url"
	"reflect"
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

type CrawlerV2CrawlSpec[R ExtracterAPIResponses] struct {
	Seeds     []string
	Extracter ExtracterAPI[R]
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

func (c *CrawlerV2) CrawlBlog(spec *CrawlerV2CrawlSpec[*ExtracterBlogResponse]) (<-chan *ExtracterBlogResponse, error) {
	resp := make(chan *ExtracterBlogResponse, 10) // TODO: config

	go func() {
		crawl[*ExtracterBlogResponse](c, spec, resp)
	}()

	return resp, nil
}

func (c *CrawlerV2) CrawlBlogSitemap(spec *CrawlerV2CrawlSpec[*ExtracterBlogSitemapResponse]) (<-chan *ExtracterBlogSitemapResponse, error) {
	resp := make(chan *ExtracterBlogSitemapResponse, 10) // TODO: config

	go func() {
		crawl[*ExtracterBlogSitemapResponse](c, spec, resp)
	}()

	return resp, nil
}

func (c *CrawlerV2) crawlThroughSiteMap(
	url string,
	extracter ExtracterAPI[*ExtracterBlogSitemapResponse],
	blogSpec *ExtracterBlogSpec,
) error {
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

		sitemapIndex, err := bodyParser(c, sitemapURL, extracter)
		if err != nil {
			multierror.Append(errors, err)
			continue
		}

		for _, link := range sitemapIndex.Links {
			// TODO: queue
			if _, err := bodyParser(c, link, NewExtracterBlog(blogSpec)); err != nil {
				c.log.Error(err)
				continue
			}
		}
	}

	return errors
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

func crawl[R ExtracterAPIResponses](c *CrawlerV2, spec *CrawlerV2CrawlSpec[R], data chan<- R) {
	for _, seed := range spec.Seeds {
		switch extracter := spec.Extracter.(type) {
		case ExtracterAPI[*ExtracterBlogResponse]:
			resp, err := bodyParser[R](c, seed, spec.Extracter)
			if err != nil {
				c.log.Error(err)
				data <- resp
				continue
			}
			data <- resp

			switch blog := extracter.(type) {
			case *ExtracterBlog:
				if blog.spec.IgnoreSiteMap {
					continue
				}

				if err := c.crawlThroughSiteMap(seed, NewExtracterBlogSitemap(), blog.spec); err != nil {
					c.log.Error(err)
					continue
				}
			}
		case ExtracterAPI[*ExtracterBlogSitemapResponse]:
			resp, err := bodyParser[R](c, seed, spec.Extracter)
			if err != nil {
				c.log.Error(err)
				data <- resp
				continue
			}

			data <- resp

			c.log.Error("strategy not implemented yet: ", reflect.TypeOf(extracter))
		}
	}
}

// TODO: pipe from goquery to extracter
func bodyParser[R ExtracterAPIResponses](c *CrawlerV2, url string, extracter ExtracterAPI[R]) (R, error) {
	resp, err := c.fetch(url)
	if err != nil {
		return nil, err
	}

	if extracter == nil {
		return nil, nil
	}

	return extracter.Extract(resp.Body) // TODO: for client-side extract get stream of new content
}
