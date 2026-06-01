package middleware

import "net/url"

type Option func(*middlewareOptions)

type middlewareOptions struct {
	sanitize func(*url.URL) string
}

func WithSanitizer(fn func(*url.URL) string) Option {
	return func(o *middlewareOptions) { o.sanitize = fn }
}

func applyOptions(opts []Option) middlewareOptions {
	var o middlewareOptions
	for _, fn := range opts {
		fn(&o)
	}
	return o
}
