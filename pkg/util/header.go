package util

import (
	"net/http"
	"strings"
)

func IsHTMLHeader(header http.Header) bool {
	return strings.Contains(header.Get("Content-Type"), "text/html")
}
