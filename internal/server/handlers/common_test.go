package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/stretchr/testify/assert"
)

type handlerSetup struct {
	handler http.Handler
	method  string
	url     string
}

type handlerTestData struct {
	testName       string
	handlerSetup   handlerSetup
	body           string
	pathParams     map[string]string
	expectedStatus int
	expectedBody   string
}

func createResponseAndRequest(data *handlerTestData) (w *httptest.ResponseRecorder, r *http.Request) {
	r = httptest.NewRequest(data.handlerSetup.method, data.handlerSetup.url, bytes.NewBufferString(data.body))
	w = httptest.NewRecorder()
	if data.pathParams != nil {
		rctx := chi.NewRouteContext()
		for key, value := range data.pathParams {
			rctx.URLParams.Add(key, value)
		}
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	}
	return w, r
}

func performHTTPHandlerTests(t *testing.T, testDatas []handlerTestData) {
	t.Helper()
	for _, testData := range testDatas {
		performHTTPHandlerTest(t, &testData)
	}
}

func performHTTPHandlerTest(t *testing.T, testData *handlerTestData) {
	t.Helper()
	t.Run(testData.testName, func(t *testing.T) {
		w, r := createResponseAndRequest(testData)

		testData.handlerSetup.handler.ServeHTTP(w, r)

		assert.Equal(t, testData.expectedStatus, w.Code, testData.testName)
		assert.Equal(t, testData.expectedBody, w.Body.String(), testData.testName)
	})
}

func performHTTPHandlerBenchmark(b *testing.B, testData *handlerTestData) {
	b.Helper()
	b.ResetTimer()

	for range b.N {
		b.StopTimer()
		w, r := createResponseAndRequest(testData)
		b.StartTimer()
		testData.handlerSetup.handler.ServeHTTP(w, r)
		b.StopTimer()
		if w.Code != testData.expectedStatus {
			b.Fail()
		}
		b.StartTimer()
	}
}
