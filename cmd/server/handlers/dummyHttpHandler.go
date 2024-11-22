package handlers

import "net/http"

type DummyHttpHandler struct{}

func (dhh *DummyHttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}
