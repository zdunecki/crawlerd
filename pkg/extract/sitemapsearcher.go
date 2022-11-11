package extract

import (
	"bytes"
	sitemap "github.com/oxffaa/gopher-parse-sitemap"
	"io"
	"net/http"
)

type SitemapSearcherResponse struct {
	Error error
	Links []string // TODO: links stream / cursor
}

func NewSitemapSearcher() API[*SitemapSearcherResponse] {
	return &SitemapSearcher{}
}

type SitemapSearcher struct {
}

// TODO: streaming
func (ebs *SitemapSearcher) Extract(reader io.Reader) (*SitemapSearcherResponse, error) {
	// TODO: use Parse and ParseIndex because it's separate method in this lib - find better solution

	var buf bytes.Buffer
	tee := io.TeeReader(reader, &buf)

	hasIndex := false

	links := make([]string, 0)

	if err := sitemap.ParseIndex(tee, func(entry sitemap.IndexEntry) error {
		hasIndex = true
		entry.GetLocation()
		resp, err := http.Get(entry.GetLocation())
		if err != nil {
			return nil
		}

		sitemap.Parse(resp.Body, func(entry sitemap.Entry) error {
			links = append(links, entry.GetLocation())
			return nil
		})

		return nil
	}); err != nil {
		return &SitemapSearcherResponse{
			Error: err,
		}, err
	}

	if !hasIndex {
		if err := sitemap.Parse(bytes.NewReader(buf.Bytes()), func(entry sitemap.Entry) error {
			links = append(links, entry.GetLocation())
			return nil
		}); err != nil {
			return &SitemapSearcherResponse{
				Error: err,
			}, err
		}
	}

	return &SitemapSearcherResponse{
		Links: links,
	}, nil
}
