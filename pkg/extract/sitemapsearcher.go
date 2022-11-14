package extract

import (
	"io"
)

type SitemapSearcherResponse struct {
	Error error
}

func NewSitemapSearcher() API[*SitemapSearcherResponse] {
	return &SitemapSearcher{}
}

type SitemapSearcher struct {
}

func (ebs *SitemapSearcher) Extract(reader io.Reader) (*SitemapSearcherResponse, error) {
	return &SitemapSearcherResponse{}, nil
}
