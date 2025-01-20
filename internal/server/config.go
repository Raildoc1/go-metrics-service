package server

import (
	"time"
)

type Config struct {
	ServerAddress   string
	SHA256Key       string
	FilePath        string
	StoreInterval   time.Duration
	ShutdownTimeout time.Duration
	NeedRestore     bool
}
