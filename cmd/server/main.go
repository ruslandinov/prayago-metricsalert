package main

import (
	"os"
	"os/signal"
	"prayago-metricsalert/internal/dbstorage"
	"prayago-metricsalert/internal/logger"
	"prayago-metricsalert/internal/memstorage"
	"prayago-metricsalert/internal/server"
	"syscall"
)

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	var ms memstorage.MemStorage
	var dbs dbstorage.DBStorager
	go func() {
		config := server.NewServerConfig()
		logger.LogSugar.Infoln("Starting server on", config.ServerAddress)

		msConfig := memstorage.MemStorageConfig{
			FPath:         config.StorageFPath,
			StoreInterval: config.StoreInterval,
			ShouldRestore: config.RestoreStorageOnStart,
		}
		ms = memstorage.NewMemStorage(msConfig)

		dbsConfig := dbstorage.DBStorageConfig{
			ConnectionString: config.DBConnectionString,
		}
		dbs = dbstorage.NewDBStorage(dbsConfig)

		if _, err := server.NewServer(config, ms, dbs); err != nil {
			logger.LogSugar.Error("Failed to start server:", err)
			logger.LogSugar.Fatalf("Failed to start server:", err)
		}
	}()

	<-stop

	logger.LogSugar.Info("Shutting down server...")
	logger.LogSugar.Info("Saving data...")
	ms.SaveData()
	dbs.Close()

	logger.LogSugar.Info("Server exiting")
}
