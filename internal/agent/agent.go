package agent

import (
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"time"
)

var needfulMemStats = [...]string{
	"Alloc",
	"BuckHashSys",
	"Frees",
	"GCCPUFraction",
	"GCSys",
	"HeapAlloc",
	"HeapIdle",
	"HeapInuse",
	"HeapObjects",
	"HeapReleased",
	"HeapSys",
	"LastGC",
	"Lookups",
	"MCacheInuse",
	"MCacheSys",
	"MSpanInuse",
	"MSpanSys",
	"Mallocs",
	"NextGC",
	"NumForcedGC",
	"NumGC",
	"OtherSys",
	"PauseTotalNs",
	"StackInuse",
	"StackSys",
	"Sys",
	"TotalAlloc",
}

type Metric struct {
	mType string
	name  string
	value string
}

type Metrics struct {
	pollCount   int64
	randomValue float64
	list        map[string]Metric
}

type AgentConfig struct {
	serverAddress  string
	serverPort     string
	pollInterval   int64
	reportInterval int64
}

type Agent struct {
	memStats runtime.MemStats
	metrics  Metrics
}

const pollCount = "PollCount"
const randomValue = "RandomValue"

func NewAgent() *Agent {
	fmt.Printf("Agent created.\r\n")

	var metrics = Metrics{
		pollCount:   0,
		randomValue: 0,
		list:        make(map[string]Metric),
	}

	return &Agent{
		metrics: metrics,
	}
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

	runtime.ReadMemStats(&agent.memStats)
	reflectedMemStats := reflect.ValueOf(agent.memStats)

	// I want it to be simple: I have the list of needed metrics, so I just loop over the MemStats
	// to get only those values I need.
	// I know that some people use JSON to bypass Go limitation on dynamic field name lookup in structs,
	// but I'm doing it in my way ;)
	for _, mName := range needfulMemStats {
		value := reflect.Indirect(reflectedMemStats).FieldByName(mName)
		// fmt.Printf("%v=%v\r\n", mName, value)

		if metric, present := agent.metrics.list[mName]; present {
			metric.value = fmt.Sprintf("%v", value)
			agent.metrics.list[mName] = metric
		} else {
			agent.metrics.list[mName] = Metric{"gauge", mName, fmt.Sprintf("%v", value)}
		}
	}

	agent.metrics.pollCount++
	agent.metrics.randomValue = rand.Float64()

	// fmt.Printf("Metrics: %v\r\n", agent.metrics)
}

func (agent *Agent) sendMetrics() {
	fmt.Printf("Agent sent metrics.\r\n")
}
