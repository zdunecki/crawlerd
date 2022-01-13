package sdk

type Option func(*options)

type options struct {
	addr string
}

func WithHTTPAddr(apiAddr string) Option {
	return func(o *options) {
		o.addr = apiAddr
	}
}
