package server

import (
	"time"
)

type Config struct {
	ServerAddress   string
	FilePath        string
	StoreInterval   time.Duration
	ShutdownTimeout time.Duration
	NeedRestore     bool
}
