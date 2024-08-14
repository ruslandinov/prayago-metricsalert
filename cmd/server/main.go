package main

import (
	"prayago-metricsalert/internal/memstorage"
	"prayago-metricsalert/internal/server"
)

func main() {
	_ = server.NewServer(memstorage.NewMemStorage(), server.NewServerConfig())
}
