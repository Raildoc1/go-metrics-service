package getallmetrics

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
)

type repository interface {
	GetAll() map[string]any
}

type handler struct {
	repository repository
	logger     Logger
}

type Logger interface {
	Error(args ...interface{})
}

func New(repository repository, logger Logger) http.Handler {
	return &handler{
		repository: repository,
		logger:     logger,
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
		h.logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, buffer.String())
	if err != nil {
		h.logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
