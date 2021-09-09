package main

type CrawlBotTestProps struct {
	RootServer string `json:"rootServer"`
}

type CrawlBotTestPage struct {
	Body           string `json:"body"`
	OutsideNetwork bool   `json:"outside_network"` // TODO: outside_network should be checked only on backend side, currently this value is set by jsx test data
}

type StringFilter struct {
	// Is apply if value is exact equal.
	Is string `json:"is,omitempty"`

	// Match apply with regular expression.
	Match string `json:"match,omitempty"`
}

type CrawlBotTestCase struct {
	StartURL           string                       `json:"start_url"`
	Description        string                       `json:"description"`
	MaxDepth           uint                         `json:"max_depth"`
	ScrapeLinksPattern string                       `json:"scrape_links_pattern"`
	FollowLinks        []*StringFilter              `json:"follow_links"`
	Pages              map[string]*CrawlBotTestPage `json:"pages"`
	Expect             []string                     `json:"expect"`
}
