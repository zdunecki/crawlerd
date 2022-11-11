package crawler

import (
	"errors"
	"fmt"
	"net/http"
)

type HTTPError struct {
	*http.Response
}

func (err *HTTPError) Error() string {
	return fmt.Sprintf("status_code:%d", err.StatusCode)
}

type Error struct {
	err error
}

func (err *Error) Error() string {
	return err.err.Error()
}

var ErrCrawlerInvalidHTTPStatus = errors.New("invalid http status")
