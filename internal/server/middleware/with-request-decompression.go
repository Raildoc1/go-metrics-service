package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
)

type compressReader struct {
	originalReader      io.ReadCloser
	decompressingReader *gzip.Reader
}

func newCompressReader(reader io.ReadCloser) (*compressReader, error) {
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		originalReader:      reader,
		decompressingReader: gzipReader,
	}, nil
}

func (r *compressReader) Read(p []byte) (n int, err error) {
	return r.decompressingReader.Read(p)
}

func (r *compressReader) Close() error {
	if err := r.originalReader.Close(); err != nil {
		return err
	}
	return r.decompressingReader.Close()
}

func WithRequestDecompression(h http.Handler, logger Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") != "gzip" {
			h.ServeHTTP(w, r)
			return
		}

		decompressingReader, err := newCompressReader(r.Body)

		if err != nil {
			switch err {
			case io.EOF: // ignore.
			default:
				logger.Errorln("failed to decompress ", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		r.Body = decompressingReader

		h.ServeHTTP(w, r)
	})
}
