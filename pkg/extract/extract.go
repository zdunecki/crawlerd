package extract

import (
	"io"
)

type Response interface {
	*SitemapResponse | *ArticleResponse
}

type API[R Response] interface {
	Extract(io.Reader) (R, error)
}
