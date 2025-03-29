package handlers

import (
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/internal/testutils"
	"net/http"
	"testing"
)

func TestGetMetric(t *testing.T) {
	serverContext := testutils.NewServerContext()
	updateMetricHandlerSetup := handlerSetup{
		handler: NewUpdateMetric(serverContext.Controller, serverContext.Logger),
		method:  http.MethodPost,
		url:     protocol.UpdateMetricURL,
	}
	getMetricHandlerSetup := handlerSetup{
		handler: NewGetMetricValue(serverContext.Repository, serverContext.Repository, serverContext.Logger),
		method:  http.MethodPost,
		url:     protocol.GetMetricURL,
	}

	tests := []handlerTestData{
		{
			testName:       "positive counter delta",
			handlerSetup:   updateMetricHandlerSetup,
			body:           testutils.TCreateCounterDeltaJSON(t, "test_counter", 1),
			expectedStatus: http.StatusOK,
		},
		{
			testName:       "get counter value",
			handlerSetup:   getMetricHandlerSetup,
			body:           testutils.TCreateCounterDeltaJSON(t, "test_counter", 1),
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":"test_counter","type":"counter","delta":1}`,
		},
		{
			testName:       "get non-existent",
			handlerSetup:   getMetricHandlerSetup,
			body:           testutils.TCreateCounterDeltaJSON(t, "non_existent", 1),
			expectedStatus: http.StatusNotFound,
		},
		{
			testName:       "set gauge",
			handlerSetup:   updateMetricHandlerSetup,
			body:           testutils.TCreateGaugeDiffJSON(t, "test_gauge", 1.3),
			expectedStatus: http.StatusOK,
		},
		{
			testName:       "get gauge value",
			handlerSetup:   getMetricHandlerSetup,
			body:           testutils.TCreateGaugeDiffJSON(t, "test_gauge", 1),
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":"test_gauge","type":"gauge","value":1.3}`,
		},
	}

	performHTTPHandlerTests(t, tests)
}

func BenchmarkGetMetric(b *testing.B) {
	serverContext := testutils.NewServerContext()
	updateMetricHandlerSetup := handlerSetup{
		handler: NewUpdateMetric(serverContext.Controller, serverContext.Logger),
		method:  http.MethodPost,
		url:     protocol.UpdateMetricURL,
	}
	getMetricHandlerSetup := handlerSetup{
		handler: NewGetMetricValue(serverContext.Repository, serverContext.Repository, serverContext.Logger),
		method:  http.MethodPost,
		url:     protocol.GetMetricURL,
	}

	updateMetricTestData := &handlerTestData{
		handlerSetup:   updateMetricHandlerSetup,
		body:           testutils.BCreateCounterDeltaJSON(b, "test_counter", 1),
		expectedStatus: http.StatusOK,
		expectedBody:   "",
	}

	w, r := createResponseAndRequest(updateMetricTestData)

	updateMetricHandlerSetup.handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		b.Fatal("benchmark setup failed")
	}

	test := handlerTestData{
		testName:       "get counter value",
		handlerSetup:   getMetricHandlerSetup,
		body:           testutils.BCreateCounterDeltaJSON(b, "test_counter", 1),
		expectedStatus: http.StatusOK,
	}

	performHTTPHandlerBenchmark(b, &test)
}
