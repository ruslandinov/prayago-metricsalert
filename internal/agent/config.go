package agent

import (
	"flag"
)

type AgentConfig struct {
	serverAddress  string
	reportInterval int64
	pollInterval   int64
}

func NewAgentConfig() AgentConfig {
	config := AgentConfig{"", 0, 0}

	flag.StringVar(&config.serverAddress, "a", "localhost:8080", "server address and port")
	flag.Int64Var(&config.reportInterval, "r", 10, "metrics sending interval")
	flag.Int64Var(&config.pollInterval, "p", 2, "metrics poll(udpate) interval")

	flag.Parse()

	return config
}
