package main

import (
	"os"
	"os/signal"
	"prayago-metricsalert/internal/logger"
	"prayago-metricsalert/internal/memstorage"
	"prayago-metricsalert/internal/server"
	"syscall"
)

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	var ms memstorage.MemStorage
	go func() {
		config := server.NewServerConfig()
		logger.LogSugar.Infoln("Starting server on", config.ServerAddress)

		msConfig := memstorage.MemStorageConfig{
			FPath:         config.StorageFPath,
			StoreInterval: config.StoreInterval,
			ShouldRestore: config.RestoreStorageOnStart,
		}
		ms = memstorage.NewMemStorage(msConfig)

		if _, err := server.NewServer(ms, config); err != nil {
			logger.LogSugar.Error("Failed to start server:", err)
			logger.LogSugar.Fatalf("Failed to start server:", err)
		}
	}()

	<-stop

	logger.LogSugar.Info("Shutting down server...")
	logger.LogSugar.Info("Saving data...")
	ms.SaveData()

	logger.LogSugar.Info("Server exiting")
}
