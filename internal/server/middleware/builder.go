package middleware

import (
	"net/http"
)

type Builder struct {
	httpHandler http.Handler
}

func NewBuilder(handler http.Handler) *Builder {
	return &Builder{
		httpHandler: handler,
	}
}

func (b *Builder) WithLogger(logger Logger) *Builder {
	b.httpHandler = WithLogger(b.httpHandler, logger)
	return b
}

func (b *Builder) WithRequestDecompression(logger Logger) *Builder {
	b.httpHandler = WithRequestDecompression(b.httpHandler, logger)
	return b
}

func (b *Builder) WithResponseCompression(logger Logger) *Builder {
	b.httpHandler = WithResponseCompression(b.httpHandler, logger)
	return b
}

func (b *Builder) Build() http.Handler {
	return b.httpHandler
}
