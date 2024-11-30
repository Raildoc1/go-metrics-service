package main

import (
	"flag"
	"go-metrics-service/cmd/common"
	"go-metrics-service/internal/server"
	"go-metrics-service/internal/server/data/storage"
	"net/http"
)

func main() {
	serverAddress := &common.ServerAddress{
		Host: "localhost",
		Port: 8080,
	}

	flag.Var(serverAddress, "a", "Server address host:port")
	flag.Parse()

	memStorage := storage.NewMemStorage()
	handler := server.NewServer(memStorage)
	err := http.ListenAndServe(serverAddress.String(), handler)
	if err != nil {
		panic(err)
	}
}
