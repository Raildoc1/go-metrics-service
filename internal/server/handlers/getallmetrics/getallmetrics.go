package getallmetrics

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
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
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, buffer.String())
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
