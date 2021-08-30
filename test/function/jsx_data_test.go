package main

import (
	"strings"
	"testing"
)

// TODO: compile every react dom with react-dom/server via golang instead of explicit in .jsx
func TestJSXData(t *testing.T) {
	var testCase []CrawlBotTestCase

	props := map[string]string{
		"fakeServer": "http://localhost:6666",
	}

	if err := jsxData("crawlbot_test.jsx", props, &testCase); err != nil {
		t.Error(err)
		return
	}

	if testCase == nil || len(testCase) == 0 {
		t.Error("test case is empty")
		return
	}

	if testCase[0].Body == "" {
		t.Error("body should be never be empty")
		return
	}

	if testCase[0].URL == "" {
		t.Error("url should be never be empty")
		return
	}

	if !strings.Contains(testCase[0].URL, props["fakeServer"]) {
		t.Error("url should contain fake server")
		return
	}
}
