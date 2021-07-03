package client

import v1 "crawlerd/api/v1"

type httpClient struct {
	url v1.V1URL
}

func (sdk *httpClient) URL() v1.V1URL {
	return sdk.url
}
