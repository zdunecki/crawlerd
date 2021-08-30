package main

type CrawlBotTestCase struct {
	URL         string            `json:"url"`
	Description string            `json:"description"`
	MaxDepth    uint              `json:"max_depth"`
	Body        string            `json:"body"`
	Pages       map[string]string `json:"pages"`
	Expect      []string          `json:"expect"`
}
