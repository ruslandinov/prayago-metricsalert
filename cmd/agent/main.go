package main

import (
	"prayago-metricsalert/internal/agent"
)

func main() {
	agent := agent.NewAgent(agent.NewAgentConfig())
	agent.Run()
}
