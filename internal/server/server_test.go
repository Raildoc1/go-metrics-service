package server

import (
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/internal/server/data/storage/memory"
	"go-metrics-service/internal/server/logging"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-resty/resty/v2"
)

func setupServer() (*httptest.Server, error) {
	memStorage := memory.NewMemStorage()
	logger, err := logging.CreateLogger(true)
	if err != nil {
		return nil, err
	}
	defer logger.Sync()
	handler := NewServer(memStorage, logger)
	return httptest.NewServer(handler), nil
}

func TestUpdate(t *testing.T) {
	server, err := setupServer()
	if err != nil {
		t.Fatal(err)
		return
	}
	defer server.Close()

	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name       string
		restPath   string
		method     string
		pathParams map[string]string
		want       want
	}{
		{
			name:     "add value",
			method:   resty.MethodPost,
			restPath: protocol.UpdateMetricValueURL,
			pathParams: map[string]string{
				protocol.TypeParam:  protocol.Counter,
				protocol.KeyParam:   "test1",
				protocol.ValueParam: "3",
			},
			want: want{
				code:        http.StatusOK,
				response:    "",
				contentType: "",
			},
		},
		{
			name:     "get value",
			method:   resty.MethodGet,
			restPath: protocol.GetMetricValueURL,
			pathParams: map[string]string{
				protocol.TypeParam: protocol.Counter,
				protocol.KeyParam:  "test1",
			},
			want: want{
				code:        http.StatusOK,
				response:    "3",
				contentType: "text/plain",
			},
		},
		{
			name:     "subtract value",
			method:   resty.MethodPost,
			restPath: protocol.UpdateMetricValueURL,
			pathParams: map[string]string{
				protocol.TypeParam:  protocol.Counter,
				protocol.KeyParam:   "test1",
				protocol.ValueParam: "-10",
			},
			want: want{
				code:        http.StatusOK,
				response:    "",
				contentType: "",
			},
		},
		{
			name:     "get value",
			method:   resty.MethodGet,
			restPath: protocol.GetMetricValueURL,
			pathParams: map[string]string{
				protocol.TypeParam: protocol.Counter,
				protocol.KeyParam:  "test1",
			},
			want: want{
				code:        http.StatusOK,
				response:    "-7",
				contentType: "text/plain",
			},
		},
		{
			name:     "get non-existing value",
			method:   resty.MethodGet,
			restPath: protocol.GetMetricValueURL,
			pathParams: map[string]string{
				protocol.TypeParam: protocol.Counter,
				protocol.KeyParam:  "test2",
			},
			want: want{
				code:        http.StatusNotFound,
				response:    "",
				contentType: "",
			},
		},
		{
			name:     "get value with wrong type",
			method:   resty.MethodGet,
			restPath: protocol.GetMetricValueURL,
			pathParams: map[string]string{
				protocol.TypeParam: protocol.Gauge,
				protocol.KeyParam:  "test1",
			},
			want: want{
				code:        http.StatusBadRequest,
				response:    "",
				contentType: "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := resty.
				New().
				SetPathParams(test.pathParams).
				R()

			url := server.URL + test.restPath

			var resp *resty.Response
			var err error

			switch test.method {
			case "GET":
				resp, err = request.Get(url)
			case "POST":
				resp, err = request.Post(url)
			default:
				require.Fail(t, "Forbidden method")
			}

			assert.NoError(t, err, "error making HTTP request")
			assert.Equal(t, test.want.code, resp.StatusCode(), "Unexpected status code")
			require.Equal(t, test.want.contentType, resp.Header().Get("Content-Type"), "Unexpected content type")
			assert.Equal(t, test.want.response, string(resp.Body()), "Unexpected response")
		})
	}
}
