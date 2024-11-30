package main

import (
	"flag"
	"go-metrics-service/cmd/common"
	"go-metrics-service/internal/agent"
	"time"
)

func main() {
	serverAddress := &common.ServerAddress{
		Host: "localhost",
		Port: 8080,
	}

	flag.Var(serverAddress, "a", "Server address host:port")
	sendingFreqSeconds := flag.Int("r", 10, "Metrics sending frequency in seconds")
	pollingFreqSeconds := flag.Int("p", 2, "Metrics polling frequency in seconds")

	flag.Parse()

	if *sendingFreqSeconds <= 0 {
		panic("sending frequency must be greater than zero")
	}

	if *pollingFreqSeconds <= 0 {
		panic("polling frequency must be greater than zero")
	}
	
	pollingFreq := time.Duration(*pollingFreqSeconds) * time.Second
	sendingFreq := time.Duration(*sendingFreqSeconds) * time.Second

	agent.Run(serverAddress.String(), pollingFreq, sendingFreq)
}
