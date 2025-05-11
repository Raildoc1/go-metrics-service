package driver

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	storagePkg "go-metrics-service/internal/agent/storage"
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/pkg/compression"
	"hash"
	"io"
	"net"
	"net/http"
	"syscall"

	"go.uber.org/zap"

	"github.com/go-resty/resty/v2"
)

var (
	ErrServerUnavailable = errors.New("server unavailable")
)

type HashFactory interface {
	Create() hash.Hash
}

type Encoder interface {
	Encode([]byte) ([]byte, error)
}

type HTTPDriver struct {
	logger      *zap.Logger
	storage     *storagePkg.Storage
	hashFactory HashFactory
	host        string
	encoder     Encoder
	ip          net.IP
}

func NewHTTPDriver(
	logger *zap.Logger,
	storage *storagePkg.Storage,
	hashFactory HashFactory,
	host string,
	encoder Encoder,
	ip net.IP,
) *HTTPDriver {
	return &HTTPDriver{
		logger:      logger,
		storage:     storage,
		hashFactory: hashFactory,
		host:        host,
		encoder:     encoder,
		ip:          ip,
	}
}

func (s *HTTPDriver) SendUpdates(ctx context.Context, metrics []protocol.Metrics) error {
	url := s.createURL(protocol.UpdateMetricsURL)
	var body bytes.Buffer
	err := compression.GzipCompress(
		metrics,
		func(writer io.Writer) compression.Encoder {
			je := json.NewEncoder(writer)
			je.SetIndent("", "")
			return je
		},
		&body,
		gzip.BestSpeed,
		s.logger,
	)
	if err != nil {
		return fmt.Errorf("failed to compress request: %w", err)
	}

	req := resty.New().
		R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("X-Real-IP", s.ip.String())

	if s.hashFactory != nil {
		h := s.hashFactory.Create()
		_, err = h.Write(body.Bytes())
		if err != nil {
			return fmt.Errorf("failed to hash: %w", err)
		}
		req = req.SetHeader(protocol.HashHeader, hex.EncodeToString(h.Sum(nil)))
	}

	bodyBytes := body.Bytes()

	if s.encoder != nil {
		encoded, err := s.encoder.Encode(bodyBytes)
		if err != nil {
			return fmt.Errorf("failed to encode request: %w", err)
		}
		bodyBytes = encoded
	}

	resp, err := req.
		SetBody(bodyBytes).
		SetLogger(NewRestyLogger(s.logger)).
		SetDebug(true).
		SetContext(ctx).
		Post(url)
	if err != nil {
		if errors.Is(err, syscall.ECONNREFUSED) {
			return fmt.Errorf("%w: %w", err, ErrServerUnavailable)
		}
		return fmt.Errorf("%w: update failed", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to send updates to %s: %d", s.host, resp.StatusCode())
	}
	return nil
}

func (s *HTTPDriver) createURL(path string) string {
	return fmt.Sprintf("http://%s%s", s.host, path)
}
