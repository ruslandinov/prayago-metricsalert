package agent

import (
	"fmt"
	"math/rand"
	"net/http"
	"prayago-metricsalert/internal/memstorage"
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
	config   AgentConfig
	memStats runtime.MemStats
	metrics  Metrics
}

const pollCount = "PollCount"
const randomValue = "RandomValue"

func NewAgent() *Agent {
	fmt.Printf("Agent created.\r\n")

	var config = AgentConfig{
		serverAddress:  "127.0.0.1",
		serverPort:     "8080",
		pollInterval:   2,
		reportInterval: 10,
	}
	var metrics = Metrics{
		pollCount:   0,
		randomValue: 0,
		list:        make(map[string]Metric),
	}

	return &Agent{
		config:  config,
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
		time.Sleep(time.Duration(agent.config.pollInterval) * time.Second)
		agent.updateMetrics()
	}
}

func (agent *Agent) startSending() {
	fmt.Printf("Agent started sending.\r\n")
	for {
		time.Sleep(time.Duration(agent.config.reportInterval) * time.Second)
		agent.sendMetrics()
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
			agent.metrics.list[mName] = Metric{memstorage.GaugeMetric, mName, fmt.Sprintf("%v", value)}
		}
	}

	agent.metrics.pollCount++
	agent.metrics.randomValue = rand.Float64()

	// fmt.Printf("Metrics: %v\r\n", agent.metrics)
}

func (agent *Agent) sendMetrics() {
	fmt.Printf("Agent sent metrics.\r\n")

	var url string
	for _, metric := range agent.metrics.list {
		url = fmt.Sprintf("http://%s:%s/update/%s/%s/%s",
			agent.config.serverAddress, agent.config.serverPort,
			metric.mType, metric.name, metric.value,
		)
		doSendMetric(url)
		// fmt.Printf("sendMetrics() url=%v\r\n", url)
	}

	// pollCount
	url = fmt.Sprintf("http://%s:%s/update/%s/%s/%v",
		agent.config.serverAddress, agent.config.serverPort,
		memstorage.CounterMetric, pollCount, agent.metrics.pollCount,
	)
	doSendMetric(url)
	// randomValue
	url = fmt.Sprintf("http://%s:%s/update/%s/%s/%v",
		agent.config.serverAddress, agent.config.serverPort,
		memstorage.GaugeMetric, randomValue, agent.metrics.randomValue,
	)
	doSendMetric(url)
}

func doSendMetric(url string) {
	// fmt.Printf("doSendMetric() url=%v\r\n", url)
	resp, err := http.Post(url, "text/plain", nil)
	if err != nil {
		fmt.Printf("doSendMetric(): url=%v, error=%v\r\n", url, err)
	} else {
		fmt.Printf("doSendMetric(): url=%v, resp=%v\r\n", url, resp)
	}
}
