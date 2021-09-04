package main

import (
	"strings"
	"testing"
)

// TODO: compile every react dom with react-dom/server via golang instead of explicit in .jsx
func TestJSXData(t *testing.T) {
	var testCase []CrawlBotTestCase

	props := &CrawlBotTestProps{
		RootServer: "http://localhost:6666",
	}

	if err := jsxData("crawlbot_test.jsx", props, &testCase); err != nil {
		t.Error(err)
		return
	}

	if testCase == nil || len(testCase) == 0 {
		t.Error("test case is empty")
		return
	}

	if testCase[0].StartURL == "" {
		t.Error("start url should be never be empty")
		return
	}

	if !strings.Contains(testCase[0].StartURL, props.RootServer) {
		t.Error("url should contain fake server")
		return
	}
}
