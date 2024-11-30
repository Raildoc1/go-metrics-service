package main

import (
	"flag"
	"go-metrics-service/cmd/common"
	"go-metrics-service/internal/server"
	"go-metrics-service/internal/server/data/storage"
)

func main() {
	serverAddress := &common.ServerAddress{
		Host: "localhost",
		Port: 8080,
	}

	flag.Var(serverAddress, "a", "Server address host:port")
	flag.Parse()

	memStorage := storage.NewMemStorage()
	err := server.Run(memStorage, serverAddress.String())
	if err != nil {
		panic(err)
	}
}
