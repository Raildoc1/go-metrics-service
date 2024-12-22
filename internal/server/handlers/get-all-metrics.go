package handlers

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"

	"go.uber.org/zap"
)

type GetAllMetricsHandler struct {
	repository AllMetricsRepository
	logger     *zap.Logger
}

func NewGetAllMetrics(repository AllMetricsRepository, logger *zap.Logger) *GetAllMetricsHandler {
	return &GetAllMetricsHandler{
		repository: repository,
		logger:     logger,
	}
}

func (h *GetAllMetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestLogger := NewRequestLogger(h.logger, r)
	defer closeBody(r.Body, requestLogger)
	data := h.repository.GetAll()
	var buffer bytes.Buffer
	for k, v := range data {
		buffer.WriteString(fmt.Sprintf("%v: %v\n", k, v))
	}
	tmpl, err := template.New("data").Parse(`{{ .}}`)
	if err != nil {
		requestLogger.Error("Failed to parse template", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "text/html")
	err = tmpl.Execute(w, buffer.String())
	if err != nil {
		requestLogger.Error("Failed to execute template", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
