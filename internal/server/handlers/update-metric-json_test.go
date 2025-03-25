package handlers

import (
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/internal/testutils"
	"net/http"
	"testing"
)

func TestUpdateMetric(t *testing.T) {
	serverContext := testutils.NewServerContext()
	handler := NewUpdateMetric(serverContext.Controller, serverContext.Logger)

	tests := []handlerTestData{
		{
			testName:       "positive counter delta",
			method:         http.MethodPost,
			url:            protocol.UpdateMetricURL,
			body:           testutils.TCreateCounterDeltaJSON(t, "test_counter", 1),
			expectedStatus: http.StatusOK,
		},
		{
			testName:       "negative counter delta",
			method:         http.MethodPost,
			url:            protocol.UpdateMetricURL,
			body:           testutils.TCreateCounterDeltaJSON(t, "test_counter", -1),
			expectedStatus: http.StatusOK,
		},
		{
			testName:       "positive gauge value",
			method:         http.MethodPost,
			url:            protocol.UpdateMetricURL,
			body:           testutils.TCreateGaugeJSON(t, "test_gauge", 1.345789),
			expectedStatus: http.StatusOK,
		},
		{
			testName:       "zero counter delta",
			method:         http.MethodPost,
			url:            protocol.UpdateMetricURL,
			body:           testutils.TCreateCounterDeltaJSON(t, "test_counter", 0),
			expectedStatus: http.StatusOK,
		},
		{
			testName:       "negative gauge value",
			method:         http.MethodPost,
			url:            protocol.UpdateMetricURL,
			body:           testutils.TCreateGaugeJSON(t, "test_gauge", -1.237859),
			expectedStatus: http.StatusOK,
		},
		{
			testName:       "update existing",
			method:         http.MethodPost,
			url:            protocol.UpdateMetricURL,
			body:           testutils.TCreateGaugeJSON(t, "test_counter", 20),
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName:       "wrong type counter",
			method:         http.MethodPost,
			url:            protocol.UpdateMetricURL,
			body:           testutils.TCreateGaugeJSON(t, "test_counter", 0),
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName:       "wrong type gauge",
			method:         http.MethodPost,
			url:            protocol.UpdateMetricURL,
			body:           testutils.TCreateCounterDeltaJSON(t, "test_gauge", 0),
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName:       "invalid JSON format",
			method:         http.MethodPost,
			url:            protocol.UpdateMetricURL,
			body:           `{"id":"test_counter""type":"counter","delta":1}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName:       "non-existent type",
			method:         http.MethodPost,
			url:            protocol.UpdateMetricURL,
			body:           `{"id":"test_counter","type":"non_existent_type","delta":1}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName:       "no value specified counter",
			method:         http.MethodPost,
			url:            protocol.UpdateMetricURL,
			body:           `{"id":"test_counter","type":"counter"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName:       "no value specified gauge",
			method:         http.MethodPost,
			url:            protocol.UpdateMetricURL,
			body:           `{"id":"test_counter","type":"gauge"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName:       "too long number",
			method:         http.MethodPost,
			url:            protocol.UpdateMetricURL,
			body:           `{"id":"test_counter","type":"counter","delta":9999999999999999999999}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName:       "wrong value type in JSON",
			method:         http.MethodPost,
			url:            protocol.UpdateMetricURL,
			body:           `{"id":"test_counter","type":"counter","value":30}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName:       "non-numeric value counter",
			method:         http.MethodPost,
			url:            protocol.UpdateMetricURL,
			body:           `{"id":"test_counter","type":"counter","delta":"hello, world!"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName:       "non-numeric value gauge",
			method:         http.MethodPost,
			url:            protocol.UpdateMetricURL,
			body:           `{"id":"test_gauge","type":"gauge","value":"hello, world!"}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	performHTTPHandlerTests(t, handler, tests)
}

func BenchmarkUpdateMetric(b *testing.B) {
	serverContext := testutils.NewServerContext()
	handler := NewUpdateMetric(serverContext.Controller, serverContext.Logger)

	test := handlerTestData{
		method:         http.MethodPost,
		url:            protocol.UpdateMetricURL,
		body:           testutils.BCreateCounterDeltaJSON(b, "test_counter", 1),
		expectedStatus: http.StatusOK,
		expectedBody:   "",
	}

	performHTTPHandlerBenchmark(b, handler, &test)
}
