package main

import (
	"prayago-metricsalert/internal/memstorage"
	"prayago-metricsalert/internal/server"
)

func main() {
	ms := memstorage.NewMemStorage()
	_ = server.NewServer(ms)
}
