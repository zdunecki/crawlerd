package main

type CrawlBotTestProps struct {
	RootServer string `json:"rootServer"`
}

type CrawlBotTestPage struct {
	Body           string `json:"body"`
	OutsideNetwork bool   `json:"outside_network"` // TODO: outside_network should be checked only on backend side, currently this value is set by jsx test data
}

type CrawlBotTestCase struct {
	StartURL    string                       `json:"start_url"`
	Description string                       `json:"description"`
	MaxDepth    uint                         `json:"max_depth"`
	Pages       map[string]*CrawlBotTestPage `json:"pages"`
	Expect      []string                     `json:"expect"`
}
