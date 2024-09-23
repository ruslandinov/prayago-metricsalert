package server

import (
	"net/http"
	"prayago-metricsalert/internal/logger"
	"prayago-metricsalert/internal/storage"
)

type Server struct {
	config  ServerConfig
	storage storage.Storage
}

func NewServer(config ServerConfig) Server {
	logger.LogSugar.Infoln("Creating server")

	storageConfig := storage.StorageConfig{
		FPath:              config.StorageFPath,
		StoreInterval:      config.StoreInterval,
		ShouldRestore:      config.RestoreStorageOnStart,
		DBConnectionString: config.DBConnectionString,
	}
	storage := storage.NewStorage(storageConfig)
	server := Server{
		config,
		storage,
	}

	logger.LogSugar.Infoln("Server created")

	return server
}

func (srv Server) StartServer() error {
	logger.LogSugar.Infoln("Starting server")
	return http.ListenAndServe(srv.config.ServerAddress, GetRouter(srv.storage))
}

func (srv Server) Stop() {
	logger.LogSugar.Infoln("Server stopping", srv.config)
	srv.storage.SaveData()
	logger.LogSugar.Infoln("Server stopped")
}
