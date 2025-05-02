package server

import (
	"go-metrics-service/internal/common/logging"
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/internal/server/controllers"
	"go-metrics-service/internal/server/data/repositories/memrepository"
	"go-metrics-service/internal/server/data/storages"
	"go-metrics-service/internal/server/data/storages/memstorage"
	"go-metrics-service/internal/server/handlers"
	"go-metrics-service/internal/server/logic"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-resty/resty/v2"
)

func setupServer() (*httptest.Server, error) {
	logger := logging.CreateZapLogger(true)
	memStorage := memstorage.New(logger)
	memRepository := memrepository.New(memStorage, logger)
	transactionManager := storages.NewDummyTransactionsManager()
	service := logic.NewService(memRepository, logger)
	controller := controllers.NewController(transactionManager, service, logger)
	mux, err := createMux(
		nil,
		memRepository,
		controller,
		make([]handlers.Pingable, 0),
		logger,
		nil,
		"",
	)
	if err != nil {
		return nil, err
	}
	return httptest.NewServer(mux), nil
}

func TestUpdate(t *testing.T) {
	server, err := setupServer()
	assert.NoError(t, err)
	defer server.Close()

	type want struct {
		response    string
		contentType string
		code        int
	}
	tests := []struct {
		name        string
		restPath    string
		method      string
		pathParams  map[string]string
		contentType string
		content     string
		want        want
	}{
		{
			name:     "add value",
			method:   resty.MethodPost,
			restPath: protocol.UpdateMetricPathParamsURL,
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
			restPath: protocol.GetMetricPathParamsURL,
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
			restPath: protocol.UpdateMetricPathParamsURL,
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
			restPath: protocol.GetMetricPathParamsURL,
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
			restPath: protocol.GetMetricPathParamsURL,
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
			restPath: protocol.GetMetricPathParamsURL,
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
		{
			name:     "get value with wrong type",
			method:   resty.MethodGet,
			restPath: protocol.GetMetricPathParamsURL,
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
		{
			name:        "update value with json",
			method:      resty.MethodPost,
			restPath:    protocol.UpdateMetricURL,
			contentType: "application/json",
			content:     `{"id":"test1","type":"counter","delta":3}`,
			want: want{
				code:        http.StatusOK,
				response:    "",
				contentType: "",
			},
		},
		{
			name:     "get value",
			method:   resty.MethodGet,
			restPath: protocol.GetMetricPathParamsURL,
			pathParams: map[string]string{
				protocol.TypeParam: protocol.Counter,
				protocol.KeyParam:  "test1",
			},
			want: want{
				code:        http.StatusOK,
				response:    "-4",
				contentType: "text/plain",
			},
		},
		{
			name:        "get value json",
			method:      resty.MethodPost,
			restPath:    protocol.GetMetricURL,
			contentType: "application/json",
			content:     `{"id":"test1","type":"counter"}`,
			want: want{
				code:        http.StatusOK,
				response:    `{"id":"test1","type":"counter","delta":-4}`,
				contentType: "application/json",
			},
		},
		{
			name:        "get non-existing value",
			method:      resty.MethodPost,
			restPath:    protocol.GetMetricURL,
			contentType: "application/json",
			content:     `{"id":"test2","type":"counter"}`,
			want: want{
				code:        http.StatusNotFound,
				response:    "",
				contentType: "",
			},
		},
		{
			name:        "get wrong type",
			method:      resty.MethodPost,
			restPath:    protocol.GetMetricURL,
			contentType: "application/json",
			content:     `{"id":"test1","type":"gauge"}`,
			want: want{
				code:        http.StatusBadRequest,
				response:    "",
				contentType: "",
			},
		},
		{
			name:        "update wrong type with json",
			method:      resty.MethodPost,
			restPath:    protocol.UpdateMetricURL,
			contentType: "application/json",
			content:     `{"id":"test1","type":"gauge","value":3}`,
			want: want{
				code:        http.StatusBadRequest,
				response:    "",
				contentType: "",
			},
		},
		{
			name:        "update wrong value with json",
			method:      resty.MethodPost,
			restPath:    protocol.UpdateMetricURL,
			contentType: "application/json",
			content:     `{"id":"test1","type":"counter","value":3}`,
			want: want{
				code:        http.StatusBadRequest,
				response:    "",
				contentType: "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := resty.New()

			if test.pathParams != nil {
				client.SetPathParams(test.pathParams)
			}

			request := client.R()

			if test.contentType != "" {
				request.Header.Add("Content-Type", test.contentType)
				request.SetBody(test.content)
			}

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
