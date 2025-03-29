package handlers

import (
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/internal/testutils"
	"net/http"
	"testing"
)

func TestUpdateMetric(t *testing.T) {
	serverContext := testutils.NewServerContext()
	updateMetricHandlerSetup := handlerSetup{
		handler: NewUpdateMetric(serverContext.Controller, serverContext.Logger),
		method:  http.MethodPost,
		url:     protocol.UpdateMetricURL,
	}

	tests := []handlerTestData{
		{
			testName:       "positive counter delta",
			handlerSetup:   updateMetricHandlerSetup,
			body:           testutils.TCreateCounterDeltaJSON(t, "test_counter", 1),
			expectedStatus: http.StatusOK,
		},
		{
			testName:       "negative counter delta",
			handlerSetup:   updateMetricHandlerSetup,
			body:           testutils.TCreateCounterDeltaJSON(t, "test_counter", -1),
			expectedStatus: http.StatusOK,
		},
		{
			testName:       "positive gauge value",
			handlerSetup:   updateMetricHandlerSetup,
			body:           testutils.TCreateGaugeDiffJSON(t, "test_gauge", 1.345789),
			expectedStatus: http.StatusOK,
		},
		{
			testName:       "zero counter delta",
			handlerSetup:   updateMetricHandlerSetup,
			body:           testutils.TCreateCounterDeltaJSON(t, "test_counter", 0),
			expectedStatus: http.StatusOK,
		},
		{
			testName:       "negative gauge value",
			handlerSetup:   updateMetricHandlerSetup,
			body:           testutils.TCreateGaugeDiffJSON(t, "test_gauge", -1.237859),
			expectedStatus: http.StatusOK,
		},
		{
			testName:       "update existing",
			handlerSetup:   updateMetricHandlerSetup,
			body:           testutils.TCreateGaugeDiffJSON(t, "test_counter", 20),
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName:       "wrong type counter",
			handlerSetup:   updateMetricHandlerSetup,
			body:           testutils.TCreateGaugeDiffJSON(t, "test_counter", 0),
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName:       "wrong type gauge",
			handlerSetup:   updateMetricHandlerSetup,
			body:           testutils.TCreateCounterDeltaJSON(t, "test_gauge", 0),
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName:       "invalid JSON format",
			handlerSetup:   updateMetricHandlerSetup,
			body:           `{"id":"test_counter""type":"counter","delta":1}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName:       "non-existent type",
			handlerSetup:   updateMetricHandlerSetup,
			body:           `{"id":"test_counter","type":"non_existent_type","delta":1}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName:       "no value specified counter",
			handlerSetup:   updateMetricHandlerSetup,
			body:           `{"id":"test_counter","type":"counter"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName:       "no value specified gauge",
			handlerSetup:   updateMetricHandlerSetup,
			body:           `{"id":"test_counter","type":"gauge"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName:       "too long number",
			handlerSetup:   updateMetricHandlerSetup,
			body:           `{"id":"test_counter","type":"counter","delta":9999999999999999999999}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName:       "wrong value type in JSON",
			handlerSetup:   updateMetricHandlerSetup,
			body:           `{"id":"test_counter","type":"counter","value":30}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName:       "non-numeric value counter",
			handlerSetup:   updateMetricHandlerSetup,
			body:           `{"id":"test_counter","type":"counter","delta":"hello, world!"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName:       "non-numeric value gauge",
			handlerSetup:   updateMetricHandlerSetup,
			body:           `{"id":"test_gauge","type":"gauge","value":"hello, world!"}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	performHTTPHandlerTests(t, tests)
}

func BenchmarkUpdateMetric(b *testing.B) {
	serverContext := testutils.NewServerContext()
	updateMetricHandlerSetup := handlerSetup{
		handler: NewUpdateMetric(serverContext.Controller, serverContext.Logger),
		method:  http.MethodPost,
		url:     protocol.UpdateMetricURL,
	}

	test := handlerTestData{
		handlerSetup:   updateMetricHandlerSetup,
		body:           testutils.BCreateCounterDeltaJSON(b, "test_counter", 1),
		expectedStatus: http.StatusOK,
	}

	performHTTPHandlerBenchmark(b, &test)
}
