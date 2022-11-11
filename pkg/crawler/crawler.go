package crawler

import (
	"crawlerd/pkg/extract"
	"github.com/hashicorp/go-multierror"
	log "github.com/sirupsen/logrus"
	"net/http"
	urllib "net/url"
	"reflect"
	"time"
)

// TODO: robots.txt, seed, sitemap (sitemap.xml, sitemap_index.xml, feed.atom)

type Spec[R extract.Response] struct {
	Seeds     []string
	Extracter extract.API[R]
}

type DefaultCrawler struct {
	http *http.Client
	log  *log.Entry
}

func New() *DefaultCrawler {
	return &DefaultCrawler{
		http: &http.Client{
			Timeout: time.Minute,
		},
		log: log.WithField("component", "crawler"),
	}
}

func (c *DefaultCrawler) CrawlBlog(spec *Spec[*extract.ArticleResponse]) (<-chan *extract.ArticleResponse, error) {
	resp := make(chan *extract.ArticleResponse, 10) // TODO: config

	go func() {
		crawl[*extract.ArticleResponse](c, spec, resp)
	}()

	return resp, nil
}

func (c *DefaultCrawler) CrawlBlogSitemap(spec *Spec[*extract.SitemapResponse]) (<-chan *extract.SitemapResponse, error) {
	resp := make(chan *extract.SitemapResponse, 10) // TODO: config

	go func() {
		crawl[*extract.SitemapResponse](c, spec, resp)
	}()

	return resp, nil
}

func (c *DefaultCrawler) crawlThroughSiteMap(
	url string,
	extracter extract.API[*extract.SitemapResponse],
	blogSpec *extract.ArticleSpec,
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
			if _, err := bodyParser(c, link, extract.NewArticle(blogSpec)); err != nil {
				c.log.Error(err)
				continue
			}
		}
	}

	return errors
}

func (c *DefaultCrawler) fetch(url string) (*http.Response, error) {
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
		return nil, multierror.Append(nil, &Error{
			err: ErrCrawlerInvalidHTTPStatus,
		}, &HTTPError{head})
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	return c.http.Do(req)
}

func crawl[R extract.Response](c *DefaultCrawler, spec *Spec[R], data chan<- R) {
	for _, seed := range spec.Seeds {
		switch extracter := spec.Extracter.(type) {

		case extract.API[*extract.ArticleResponse]:
			resp, err := bodyParser[R](c, seed, spec.Extracter)
			if err != nil {
				c.log.Error(err)
				data <- resp
				continue
			}
			data <- resp

			switch article := extracter.(type) {

			case *extract.Article:
				if article.Spec.DeepLink {
					// TODO: other techniques than sitemap
					if err := c.crawlThroughSiteMap(seed, extract.NewSitemap(), article.Spec); err != nil {
						c.log.Error(err)
						continue
					}
				}
			}

		case extract.API[*extract.SitemapResponse]:
			resp, err := bodyParser[R](c, seed, spec.Extracter)
			if err != nil {
				c.log.Error(err)
				data <- resp
				continue
			}

			data <- resp

		default:
			c.log.Error("strategy not implemented yet: ", reflect.TypeOf(extracter))
		}
	}
}

// TODO: pipe from goquery to extracter
func bodyParser[R extract.Response](c *DefaultCrawler, url string, extracter extract.API[R]) (R, error) {
	resp, err := c.fetch(url)
	if err != nil {
		return nil, err
	}

	if extracter == nil {
		return nil, nil
	}

	return extracter.Extract(resp.Body) // TODO: for client-side extract get stream of new content
}
