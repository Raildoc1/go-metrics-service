package handlers

import (
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/internal/testutils"
	"net/http"
	"testing"
)

func TestUpdateMetricPathParams(t *testing.T) {
	serverContext := testutils.NewServerContext()
	updateMetricHandlerSetup := handlerSetup{
		handler: NewUpdateMetricPathParams(serverContext.Controller, serverContext.Logger),
		method:  http.MethodPost,
		url:     protocol.UpdateMetricPathParamsURL,
	}

	tests := []handlerTestData{
		{
			testName:     "positive counter delta",
			handlerSetup: updateMetricHandlerSetup,
			pathParams: map[string]string{
				protocol.TypeParam:  protocol.Counter,
				protocol.KeyParam:   "test_counter",
				protocol.ValueParam: "1",
			},
			expectedStatus: http.StatusOK,
		},
		{
			testName:     "negative counter delta",
			handlerSetup: updateMetricHandlerSetup,
			pathParams: map[string]string{
				protocol.TypeParam:  protocol.Counter,
				protocol.KeyParam:   "test_counter",
				protocol.ValueParam: "-1",
			},
			expectedStatus: http.StatusOK,
		},
		{
			testName:     "zero counter delta",
			handlerSetup: updateMetricHandlerSetup,
			pathParams: map[string]string{
				protocol.TypeParam:  protocol.Counter,
				protocol.KeyParam:   "test_counter",
				protocol.ValueParam: "0",
			},
			expectedStatus: http.StatusOK,
		},
		{
			testName:     "positive gauge value",
			handlerSetup: updateMetricHandlerSetup,
			pathParams: map[string]string{
				protocol.TypeParam:  protocol.Gauge,
				protocol.KeyParam:   "test_gauge",
				protocol.ValueParam: "1.345789",
			},
			expectedStatus: http.StatusOK,
		},
		{
			testName:     "negative gauge value",
			handlerSetup: updateMetricHandlerSetup,
			pathParams: map[string]string{
				protocol.TypeParam:  protocol.Gauge,
				protocol.KeyParam:   "test_gauge",
				protocol.ValueParam: "-1.237859",
			},
			expectedStatus: http.StatusOK,
		},
		{
			testName:     "update existing",
			handlerSetup: updateMetricHandlerSetup,
			pathParams: map[string]string{
				protocol.TypeParam:  protocol.Counter,
				protocol.KeyParam:   "test_counter",
				protocol.ValueParam: "20",
			},
			expectedStatus: http.StatusOK,
		},
		{
			testName:     "wrong type counter",
			handlerSetup: updateMetricHandlerSetup,
			pathParams: map[string]string{
				protocol.TypeParam:  protocol.Gauge,
				protocol.KeyParam:   "test_counter",
				protocol.ValueParam: "1.23478",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName:     "wrong type gauge",
			handlerSetup: updateMetricHandlerSetup,
			pathParams: map[string]string{
				protocol.TypeParam:  protocol.Counter,
				protocol.KeyParam:   "test_gauge",
				protocol.ValueParam: "30",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName:     "wrong value type",
			handlerSetup: updateMetricHandlerSetup,
			pathParams: map[string]string{
				protocol.TypeParam:  protocol.Counter,
				protocol.KeyParam:   "test_counter",
				protocol.ValueParam: "4.345",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName:     "non-existent type",
			handlerSetup: updateMetricHandlerSetup,
			pathParams: map[string]string{
				protocol.TypeParam:  "non_existent_type",
				protocol.KeyParam:   "test_counter",
				protocol.ValueParam: "4.345",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName:     "too long number",
			handlerSetup: updateMetricHandlerSetup,
			pathParams: map[string]string{
				protocol.TypeParam:  protocol.Counter,
				protocol.KeyParam:   "test_counter",
				protocol.ValueParam: "9999999999999999999999",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName:     "non-numeric value counter",
			handlerSetup: updateMetricHandlerSetup,
			pathParams: map[string]string{
				protocol.TypeParam:  protocol.Counter,
				protocol.KeyParam:   "test_counter",
				protocol.ValueParam: "hello, world!",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName:     "non-numeric value gauge",
			handlerSetup: updateMetricHandlerSetup,
			pathParams: map[string]string{
				protocol.TypeParam:  protocol.Gauge,
				protocol.KeyParam:   "test_gauge",
				protocol.ValueParam: "hello, world!",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	performHTTPHandlerTests(t, tests)
}

func BenchmarkUpdateMetricPathParams(b *testing.B) {
	serverContext := testutils.NewServerContext()
	updateMetricHandlerSetup := handlerSetup{
		handler: NewUpdateMetricPathParams(serverContext.Controller, serverContext.Logger),
		method:  http.MethodPost,
		url:     protocol.UpdateMetricPathParamsURL,
	}

	test := handlerTestData{
		testName:     "positive counter delta",
		handlerSetup: updateMetricHandlerSetup,
		pathParams: map[string]string{
			protocol.TypeParam:  protocol.Counter,
			protocol.KeyParam:   "test_counter",
			protocol.ValueParam: "1",
		},
		expectedStatus: http.StatusOK,
	}

	performHTTPHandlerBenchmark(b, &test)
}
