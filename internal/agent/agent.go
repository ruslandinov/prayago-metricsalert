package agent

import (
	"fmt"
	"time"
)

type AgentConfig struct {
	serverAddress  string
	serverPort     string
	pollInterval   int64
	reportInterval int64
}

type Agent struct {
}

func NewAgent() *Agent {
	fmt.Printf("Agent created.\r\n")

	return &Agent{}
}

func (agent *Agent) Run() {
	fmt.Printf("Agent started.\r\n")
	go agent.startPolling()
	go agent.startSending()
	for {
		time.Sleep(1 * time.Second)
	}
}

func (agent *Agent) startPolling() {
	fmt.Printf("Agent started polling.\r\n")
	for {
		agent.updateMetrics()
		time.Sleep(2 * time.Second)
	}
}

func (agent *Agent) startSending() {
	fmt.Printf("Agent started sending.\r\n")
	for {
		agent.sendMetrics()
		time.Sleep(10 * time.Second)
	}
}

func (agent *Agent) updateMetrics() {
	fmt.Printf("Agent updated metrics.\r\n")
}

func (agent *Agent) sendMetrics() {
	fmt.Printf("Agent sent metrics.\r\n")
}
