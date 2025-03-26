package handlers

import (
	"go-metrics-service/internal/common/protocol"
	"go-metrics-service/internal/testutils"
	"net/http"
	"testing"
)

func TestUpdateMetrics(t *testing.T) {
	serverContext := testutils.NewServerContext()
	updateMetricHandlerSetup := handlerSetup{
		handler: NewUpdateMetrics(serverContext.Controller, serverContext.Logger),
		method:  http.MethodPost,
		url:     protocol.UpdateMetricsURL,
	}

	tests := []handlerTestData{
		{
			testName:     "simple update",
			handlerSetup: updateMetricHandlerSetup,
			body: testutils.TCreateMetricsJSON(t, []protocol.Metrics{
				testutils.CreateCounter("test_counter", 1),
				testutils.CreateGauge("test_gauge", 1.4),
			}),
			expectedStatus: http.StatusOK,
		},
		{
			testName:       "empty body",
			handlerSetup:   updateMetricHandlerSetup,
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName:       "invalid JSON",
			handlerSetup:   updateMetricHandlerSetup,
			body:           `[{id: "test_counter", ,}]`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName:     "type change in one request",
			handlerSetup: updateMetricHandlerSetup,
			body: testutils.TCreateMetricsJSON(t, []protocol.Metrics{
				testutils.CreateCounter("new_test_counter", 1),
				testutils.CreateGauge("new_test_counter", 1.4),
			}),
			expectedStatus: http.StatusBadRequest,
		},
	}

	performHTTPHandlerTests(t, tests)
}

func BenchmarkUpdateMetrics(b *testing.B) {
	serverContext := testutils.NewServerContext()
	updateMetricsHandlerSetup := handlerSetup{
		handler: NewUpdateMetrics(serverContext.Controller, serverContext.Logger),
		method:  http.MethodPost,
		url:     protocol.UpdateMetricsURL,
	}

	test := handlerTestData{
		handlerSetup: updateMetricsHandlerSetup,
		body: testutils.BCreateMetricsJSON(b, []protocol.Metrics{
			testutils.CreateCounter("test_counter", 1),
			testutils.CreateGauge("test_gauge", 1.4),
		}),
		expectedStatus: http.StatusOK,
	}

	performHTTPHandlerBenchmark(b, &test)
}
