package handlers

import "net/http"

type DummyHTTPHandler struct{}

func (dhh *DummyHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}
