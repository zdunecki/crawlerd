package extract

import (
	"io"
)

type Response interface {
	*SitemapResponse | *ArticleResponse | *SitemapSearcherResponse
}

type API[R Response] interface {
	Extract(io.Reader) (R, error)
}
