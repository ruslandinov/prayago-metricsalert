package server

import (
	"flag"
)

type ServerConfig struct {
	serverAddress string
}

func NewServerConfig() ServerConfig {
	config := ServerConfig{""}
	flag.StringVar(&config.serverAddress, "a", "localhost:8080", "server address and port")
	flag.Parse()

	return config
}
