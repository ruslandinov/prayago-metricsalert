package agent

import (
	"flag"
	"fmt"
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
	a := flag.String("a", "localhost:8080", "server address and port")
	r := flag.Int("r", 10, "metrics sending interval")
	p := flag.Int("p", 2, "metrics poll(udpate) interval")
	flag.Parse()

	config := AgentConfig{*a, time.Duration(*r) * time.Second, time.Duration(*p) * time.Second}

	if envServerAddress := os.Getenv("ADDRESS"); envServerAddress != "" {
		config.serverAddress = envServerAddress
	}
	if envReportIntrvl := os.Getenv("REPORT_INTERVAL"); envReportIntrvl != "" {
		if reportIntervalInt, err := strconv.Atoi(envReportIntrvl); err == nil {
			config.reportInterval = time.Duration(reportIntervalInt)
		}
	}
	if envPollIntrvl := os.Getenv("POLL_INTERVAL"); envPollIntrvl != "" {
		if pollIntervalInt, err := strconv.Atoi(envPollIntrvl); err == nil {
			config.pollInterval = time.Duration(pollIntervalInt)
		}
	}

	return config
}
