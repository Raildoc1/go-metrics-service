package middleware

import (
	"net/http"

	"go.uber.org/zap"
)

type Builder struct {
	httpHandler http.Handler
}

func NewBuilder(handler http.Handler) *Builder {
	return &Builder{
		httpHandler: handler,
	}
}

func (b *Builder) WithLogger(logger *zap.Logger) *Builder {
	b.httpHandler = withLogger(b.httpHandler, logger)
	return b
}

func (b *Builder) WithRequestDecompression(logger *zap.Logger) *Builder {
	b.httpHandler = withRequestDecompression(b.httpHandler, logger)
	return b
}

func (b *Builder) WithResponseCompression(logger *zap.Logger) *Builder {
	b.httpHandler = withResponseCompression(b.httpHandler, logger)
	return b
}

func (b *Builder) Build() http.Handler {
	return b.httpHandler
}
