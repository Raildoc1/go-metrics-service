package testutils

import (
	"go-metrics-service/internal/server/controllers"
	"go-metrics-service/internal/server/data/repositories/memrepository"
	"go-metrics-service/internal/server/data/storages"
	"go-metrics-service/internal/server/data/storages/memstorage"
	"go-metrics-service/internal/server/logic"

	"go.uber.org/zap"
)

type ServerContext struct {
	Repository logic.Repository
	Controller *controllers.Controller
	Logger     *zap.Logger
}

func NewServerContext() *ServerContext {
	logger := zap.NewNop()
	memStorage := memstorage.New(logger)
	memRepository := memrepository.New(memStorage, logger)
	transactionManager := storages.NewDummyTransactionsManager()
	service := logic.NewService(memRepository, logger)
	controller := controllers.NewController(transactionManager, service, logger)
	return &ServerContext{
		Repository: memRepository,
		Controller: controller,
		Logger:     logger,
	}
}
