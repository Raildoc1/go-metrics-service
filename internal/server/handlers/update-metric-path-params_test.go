package handlers

import (
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/internal/testutils"
	"net/http"
	"testing"
)

func TestUpdateMetricPathParams(t *testing.T) {
	serverContext := testutils.NewServerContext()
	handler := NewUpdateMetricPathParams(serverContext.Controller, serverContext.Logger)

	tests := []handlerTestData{
		{
			testName: "positive counter delta",
			method:   http.MethodPost,
			url:      protocol.UpdateMetricPathParamsURL,
			pathParams: map[string]string{
				protocol.TypeParam:  protocol.Counter,
				protocol.KeyParam:   "test_counter",
				protocol.ValueParam: "1",
			},
			expectedStatus: http.StatusOK,
		},
		{
			testName: "negative counter delta",
			method:   http.MethodPost,
			url:      protocol.UpdateMetricPathParamsURL,
			pathParams: map[string]string{
				protocol.TypeParam:  protocol.Counter,
				protocol.KeyParam:   "test_counter",
				protocol.ValueParam: "-1",
			},
			expectedStatus: http.StatusOK,
		},
		{
			testName: "zero counter delta",
			method:   http.MethodPost,
			url:      protocol.UpdateMetricPathParamsURL,
			pathParams: map[string]string{
				protocol.TypeParam:  protocol.Counter,
				protocol.KeyParam:   "test_counter",
				protocol.ValueParam: "0",
			},
			expectedStatus: http.StatusOK,
		},
		{
			testName: "positive gauge value",
			method:   http.MethodPost,
			url:      protocol.UpdateMetricPathParamsURL,
			pathParams: map[string]string{
				protocol.TypeParam:  protocol.Gauge,
				protocol.KeyParam:   "test_gauge",
				protocol.ValueParam: "1.345789",
			},
			expectedStatus: http.StatusOK,
		},
		{
			testName: "negative gauge value",
			method:   http.MethodPost,
			url:      protocol.UpdateMetricPathParamsURL,
			pathParams: map[string]string{
				protocol.TypeParam:  protocol.Gauge,
				protocol.KeyParam:   "test_gauge",
				protocol.ValueParam: "-1.237859",
			},
			expectedStatus: http.StatusOK,
		},
		{
			testName: "update existing",
			method:   http.MethodPost,
			url:      protocol.UpdateMetricPathParamsURL,
			pathParams: map[string]string{
				protocol.TypeParam:  protocol.Counter,
				protocol.KeyParam:   "test_counter",
				protocol.ValueParam: "20",
			},
			expectedStatus: http.StatusOK,
		},
		{
			testName: "wrong type counter",
			method:   http.MethodPost,
			url:      protocol.UpdateMetricPathParamsURL,
			pathParams: map[string]string{
				protocol.TypeParam:  protocol.Gauge,
				protocol.KeyParam:   "test_counter",
				protocol.ValueParam: "1.23478",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName: "wrong type gauge",
			method:   http.MethodPost,
			url:      protocol.UpdateMetricPathParamsURL,
			pathParams: map[string]string{
				protocol.TypeParam:  protocol.Counter,
				protocol.KeyParam:   "test_gauge",
				protocol.ValueParam: "30",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName: "wrong value type",
			method:   http.MethodPost,
			url:      protocol.UpdateMetricPathParamsURL,
			pathParams: map[string]string{
				protocol.TypeParam:  protocol.Counter,
				protocol.KeyParam:   "test_counter",
				protocol.ValueParam: "4.345",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName: "non-existent type",
			method:   http.MethodPost,
			url:      protocol.UpdateMetricPathParamsURL,
			pathParams: map[string]string{
				protocol.TypeParam:  "non_existent_type",
				protocol.KeyParam:   "test_counter",
				protocol.ValueParam: "4.345",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName: "too long number",
			method:   http.MethodPost,
			url:      protocol.UpdateMetricPathParamsURL,
			pathParams: map[string]string{
				protocol.TypeParam:  protocol.Counter,
				protocol.KeyParam:   "test_counter",
				protocol.ValueParam: "9999999999999999999999",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName: "non-numeric value counter",
			method:   http.MethodPost,
			url:      protocol.UpdateMetricPathParamsURL,
			pathParams: map[string]string{
				protocol.TypeParam:  protocol.Counter,
				protocol.KeyParam:   "test_counter",
				protocol.ValueParam: "hello, world!",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName: "non-numeric value gauge",
			method:   http.MethodPost,
			url:      protocol.UpdateMetricPathParamsURL,
			pathParams: map[string]string{
				protocol.TypeParam:  protocol.Gauge,
				protocol.KeyParam:   "test_gauge",
				protocol.ValueParam: "hello, world!",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	performHTTPHandlerTests(t, handler, tests)
}

func BenchmarkUpdateMetricPathParams(b *testing.B) {
	serverContext := testutils.NewServerContext()
	handler := NewUpdateMetric(serverContext.Controller, serverContext.Logger)

	test := handlerTestData{
		testName:       "positive counter delta",
		method:         http.MethodPost,
		url:            "/update/counter/test_counter/1",
		expectedStatus: http.StatusOK,
	}

	performHTTPHandlerBenchmark(b, handler, &test)
}
