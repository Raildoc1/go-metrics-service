package server

import "time"

type Config struct {
	ServerAddress string
	NeedRestore   bool
	FilePath      string
	StoreInterval time.Duration
}
