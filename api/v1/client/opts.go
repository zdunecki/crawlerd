package client

type Option func(*options)

type options struct {
	addr string
}

func WithHTTP(apiURL string) Option {
	return func(o *options) {
		o.addr = apiURL
	}
}
