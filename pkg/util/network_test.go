package util

import (
	"testing"

	"crawlerd/test"
)

// TODO: util should not be in pkg

func TestAbc(t *testing.T) {
	testCases := []struct {
		input  string
		expect string
	}{
		{
			input:  "http://localhost:8080/v1",
			expect: "http://localhost:8080",
		},
		{
			input:  "http://localhost:8080",
			expect: "http://localhost:8080",
		},
		{
			input:  "http://localhost:8080/long/long",
			expect: "http://localhost:8080",
		},
		{
			input:  "https://localhost:8080/v1",
			expect: "https://localhost:8080",
		},
		{
			input:  "https://localhost",
			expect: "https://localhost",
		},
		{
			input:  "https://example.com",
			expect: "https://example.com",
		},
		{
			input:  "https://example.com:8080",
			expect: "https://example.com:8080",
		},
		{
			input:  "https://example.com:8080/v1",
			expect: "https://example.com:8080",
		},
	}

	for _, tc := range testCases {
		result := BaseAddr(tc.input)
		test.Diff(t, "should be equal", tc.expect, result)
	}
}
