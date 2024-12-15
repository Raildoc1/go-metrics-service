package handlers

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
)

type GetAllMetricsHandler struct {
	repository AllMetricsRepository
	logger     Logger
}

func NewGetAllMetrics(repository AllMetricsRepository, logger Logger) *GetAllMetricsHandler {
	return &GetAllMetricsHandler{
		repository: repository,
		logger:     logger,
	}
}

func (h *GetAllMetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data := h.repository.GetAll()
	var buffer bytes.Buffer
	for k, v := range data {
		buffer.WriteString(fmt.Sprintf("%v: %v\n", k, v))
	}
	tmpl, err := template.New("data").Parse(`{{ .}}`)
	if err != nil {
		h.logger.Errorln(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, buffer.String())
	if err != nil {
		h.logger.Errorln(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
}
