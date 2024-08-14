package agent

import (
	"flag"
	"os"
	"strconv"
	"time"
)

type AgentConfig struct {
	serverAddress  string
	reportInterval time.Duration
	pollInterval   time.Duration
}

func NewAgentConfig() AgentConfig {
	serverAddress := *flag.String("a", "localhost:8080", "server address and port")
	reportInterval := time.Duration(*flag.Int("r", 10, "metrics sending interval"))
	pollInterval := time.Duration(*flag.Int("p", 2, "metrics poll(udpate) interval"))
	flag.Parse()

	if envServerAddress := os.Getenv("ADDRESS"); envServerAddress != "" {
		serverAddress = envServerAddress
	}
	if envReportIntrvl := os.Getenv("REPORT_INTERVAL"); envReportIntrvl != "" {
		if reportIntervalInt, err := strconv.Atoi(envReportIntrvl); err == nil {
			reportInterval = time.Duration(reportIntervalInt)
		}
	}
	if envPollIntrvl := os.Getenv("POLL_INTERVAL"); envPollIntrvl != "" {
		if pollIntervalInt, err := strconv.Atoi(envPollIntrvl); err == nil {
			pollInterval = time.Duration(pollIntervalInt)
		}
	}

	return AgentConfig{serverAddress, reportInterval, pollInterval}
}
