package main

import (
	// "fmt"
	"os"
	"os/signal"
	"prayago-metricsalert/internal/logger"
	"prayago-metricsalert/internal/server"
	"syscall"
	"time"
)

func main() {
	logger.LogSugar.Infoln("Starting server")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	var srvr server.Server
	go func() {
		config := server.NewServerConfig()
		srvr = server.NewServer(config)
		srvr.StartServer()
		// logger.LogSugar.Error("Failed to start server:", err)
		// logger.LogSugar.Fatalf("Failed to start server:", err)
	}()

	<-stop

	logger.LogSugar.Infoln("Stoping server")
	srvr.Stop()

	time.Sleep(2 * time.Second)
	logger.LogSugar.Infoln("Server exit")
}
