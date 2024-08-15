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

type Agent struct {
	config   AgentConfig
	memStats runtime.MemStats
	metrics  Metrics
}

const pollCount = "PollCount"
const randomValue = "RandomValue"

func NewAgent(config AgentConfig) *Agent {
	fmt.Printf("Agent created.\r\n")

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
	fmt.Printf("Agent started polling. %v\r\n", agent.config.pollInterval)
	for {
		time.Sleep(agent.config.pollInterval)
		agent.updateMetrics()
	}
}

func (agent *Agent) startSending() {
	fmt.Printf("Agent started sending. %v\r\n", agent.config.reportInterval)
	for {
		time.Sleep(agent.config.reportInterval)
		agent.sendMetrics()
	}
}

func (agent *Agent) updateMetrics() {
	// fmt.Printf("Agent updated metrics.\r\n")

	runtime.ReadMemStats(&agent.memStats)
	reflectedMemStats := reflect.ValueOf(agent.memStats)

	// I want it to be simple: I have the list of needed metrics, so I just loop over the MemStats
	// to get only those values I need.
	// I know that some people use JSON to bypass Go limitation on dynamic field name lookup in structs,
	// but I'm doing it in my way ;)
	for _, mName := range needfulMemStats {
		value := reflect.Indirect(reflectedMemStats).FieldByName(mName)
		if metric, present := agent.metrics.list[mName]; present {
			metric.value = fmt.Sprintf("%v", value)
			agent.metrics.list[mName] = metric
		} else {
			agent.metrics.list[mName] = Metric{memstorage.GaugeMetric, mName, fmt.Sprintf("%v", value)}
		}
	}

	agent.metrics.pollCount++
	agent.metrics.randomValue = rand.Float64()
}

func (agent *Agent) sendMetrics() {
	// fmt.Printf("Agent sent metrics.\r\n")

	var url string
	for _, metric := range agent.metrics.list {
		url = fmt.Sprintf("http://%s/update/%s/%s/%s",
			agent.config.serverAddress,
			metric.mType, metric.name, metric.value,
		)
		doSendMetric(url)
		// fmt.Printf("sendMetrics() url=%v\r\n", url)
	}

	// pollCount
	url = fmt.Sprintf("http://%s/update/%s/%s/%v",
		agent.config.serverAddress,
		memstorage.CounterMetric, pollCount, agent.metrics.pollCount,
	)
	doSendMetric(url)
	// randomValue
	url = fmt.Sprintf("http://%s/update/%s/%s/%v",
		agent.config.serverAddress,
		memstorage.GaugeMetric, randomValue, agent.metrics.randomValue,
	)
	doSendMetric(url)
}

func doSendMetric(url string) {
	// fmt.Printf("doSendMetric() url=%v\r\n", url)
	resp, err := http.Post(url, "text/plain", nil)
	if err != nil {
		// fmt.Printf("doSendMetric(): url=%v, error=%v\r\n", url, err)
		return
	}
	defer resp.Body.Close()
	// fmt.Printf("doSendMetric(): url=%v, resp=%v\r\n", url, resp)
}
