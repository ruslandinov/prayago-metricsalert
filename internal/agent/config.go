package agent

import (
	"flag"
	"time"
)

type AgentConfig struct {
	serverAddress  string
	reportInterval time.Duration
	pollInterval   time.Duration
}

func NewAgentConfig() AgentConfig {
	serverAddress := flag.String("a", "localhost:8080", "server address and port")
	reportInterval := time.Duration(*flag.Int("r", 10, "metrics sending interval"))
	pollInterval := time.Duration(*flag.Int("p", 2, "metrics poll(udpate) interval"))
	flag.Parse()

	return AgentConfig{*serverAddress, reportInterval, pollInterval}
}
