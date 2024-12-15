package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"slices"
	"strings"
)

const (
	contentApplicationJSON = "application/json"
	contentTextPlain       = "text/plain"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func WithResponseCompression(h http.Handler, logger Logger) http.Handler {
	contentToCompress := []string{
		contentTextPlain,
		contentApplicationJSON,
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			logger.Debugln(r.Method, " ", r.URL, " compression missed")
			h.ServeHTTP(w, r)
			return
		}
		if !slices.Contains(contentToCompress, r.Header.Get("Content-Type")) {
			h.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			logger.Errorln(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func(gz *gzip.Writer) {
			err := gz.Close()
			if err != nil {
				logger.Errorln(err)
				return
			}
		}(gz)

		w.Header().Set("Content-Encoding", "gzip")
		h.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
		logger.Debugln(r.Method, " ", r.URL, " request compressed")
	})
}
