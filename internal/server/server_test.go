package server

import (
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/internal/server/data/storage"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupServer() *httptest.Server {
	memStorage := storage.NewMemStorage()
	handler := NewServer(memStorage)
	return httptest.NewServer(handler)
}

func TestUpdate(t *testing.T) {
	server := setupServer()
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
			method:   "POST",
			restPath: protocol.UpdateMetricValueUrl,
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
			method:   "GET",
			restPath: protocol.GetMetricValueUrl,
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
			method:   "POST",
			restPath: protocol.UpdateMetricValueUrl,
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
			method:   "GET",
			restPath: protocol.GetMetricValueUrl,
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
			method:   "GET",
			restPath: protocol.GetMetricValueUrl,
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
			method:   "GET",
			restPath: protocol.GetMetricValueUrl,
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
	}
}
