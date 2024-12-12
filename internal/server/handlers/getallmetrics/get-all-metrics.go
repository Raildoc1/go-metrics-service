package getallmetrics

import (
	"bytes"
	"fmt"
	"go-metrics-service/internal/server/logger"
	"html/template"
	"net/http"
)

type repository interface {
	GetAll() map[string]any
}

type handler struct {
	repository repository
}

func New(repository repository) http.Handler {
	return &handler{
		repository: repository,
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data := h.repository.GetAll()
	var buffer bytes.Buffer
	for k, v := range data {
		buffer.WriteString(fmt.Sprintf("%v: %v\n", k, v))
	}
	tmpl, err := template.New("data").Parse(`{{ .}}`)
	if err != nil {
		logger.Log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, buffer.String())
	if err != nil {
		logger.Log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
