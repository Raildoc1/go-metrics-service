package server

import (
	"go-metrics-service/internal/server/database"
	"time"
)

type Config struct {
	Database        database.Config
	ServerAddress   string
	FilePath        string
	StoreInterval   time.Duration
	ShutdownTimeout time.Duration
	NeedRestore     bool
}
