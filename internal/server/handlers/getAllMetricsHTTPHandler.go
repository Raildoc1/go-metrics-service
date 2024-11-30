package handlers

import (
	"bytes"
	"fmt"
	"go-metrics-service/internal/server/data/repositories"
	"html/template"
	"net/http"
)

type getAllMetricsHTTPHandler struct {
	storage repositories.Storage
}

func NewGetAllMetricsHTTPHandler(
	storage repositories.Storage,
) http.Handler {
	return &getAllMetricsHTTPHandler{
		storage: storage,
	}
}

func (h *getAllMetricsHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data := h.storage.GetAll()
	var buffer bytes.Buffer
	for k, v := range data {
		buffer.WriteString(fmt.Sprintf("%v: %v\n", k, v))
	}
	tmpl, err := template.New("data").Parse(`{{ .}}`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, buffer.String())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
