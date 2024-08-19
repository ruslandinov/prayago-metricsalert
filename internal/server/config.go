package server

import (
	"flag"
	"os"
)

type ServerConfig struct {
	serverAddress string
}

func NewServerConfig() ServerConfig {
	config := ServerConfig{""}
	flag.StringVar(&config.serverAddress, "a", "localhost:8080", "server address and port")
	flag.Parse()

	if envServerAddress := os.Getenv("ADDRESS"); envServerAddress != "" {
		config.serverAddress = envServerAddress
	}

	return config
}
