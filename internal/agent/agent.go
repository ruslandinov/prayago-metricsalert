package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"prayago-metricsalert/internal/memstorage"
	"prayago-metricsalert/internal/protocol"
	"reflect"
	"runtime"
	"strconv"
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

type Metric = protocol.Metric

type Agent struct {
	config      AgentConfig
	memStats    runtime.MemStats
	metrics     map[string]Metric
	pollCount   Metric
	randomValue Metric
}

const pollCount = "PollCount"
const randomValue = "RandomValue"

var serverJSONPOSTUpdateURI string

func NewAgent(config AgentConfig) *Agent {
	fmt.Printf("Agent created.\r\n")

	serverJSONPOSTUpdateURI = fmt.Sprintf("http://%s/update/", config.serverAddress)

	var zeroInt int64 = 0
	var zeroFloat float64 = 0
	return &Agent{
		config:      config,
		metrics:     make(map[string]Metric),
		pollCount:   Metric{ID: "PollCount", MType: memstorage.CounterMetric, Delta: &zeroInt},
		randomValue: Metric{ID: "RandomValue", MType: memstorage.GaugeMetric, Value: &zeroFloat},
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
		agent.sendJSONMetrics()
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
		valueFloat64, _ := strconv.ParseFloat(fmt.Sprintf("%v", value), 64)
		if metric, present := agent.metrics[mName]; present {
			metric.Value = &valueFloat64
			agent.metrics[mName] = metric
		} else {
			agent.metrics[mName] = Metric{ID: mName, MType: memstorage.GaugeMetric, Value: &valueFloat64}
		}
	}

	*agent.pollCount.Delta++
	*agent.randomValue.Value = rand.Float64()
	// fmt.Printf("%v \r\n\r\n", agent.metrics)
}

func (agent *Agent) sendMetrics() {
	// fmt.Printf("Agent sent metrics.\r\n")

	var url string
	for _, metric := range agent.metrics {
		url = fmt.Sprintf("http://%s/update/%s/%s/%v",
			agent.config.serverAddress,
			metric.MType, metric.ID, *metric.Value,
		)
		doSendMetric(url)
		// fmt.Printf("sendMetrics() url=%v\r\n", url)
	}

	// pollCount
	url = fmt.Sprintf("http://%s/update/%s/%s/%v",
		agent.config.serverAddress,
		memstorage.CounterMetric, pollCount, *agent.pollCount.Delta,
	)
	doSendMetric(url)
	// randomValue
	url = fmt.Sprintf("http://%s/update/%s/%s/%v",
		agent.config.serverAddress,
		memstorage.GaugeMetric, randomValue, *agent.randomValue.Value,
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

func (agent *Agent) sendJSONMetrics() {
	// fmt.Printf("Agent sent metrics.\r\n")

	for _, metric := range agent.metrics {
		doSendJSONMetric(metric)
	}

	doSendJSONMetric(agent.pollCount)
	doSendJSONMetric(agent.randomValue)
}

func doSendJSONMetric(metric Metric) {
	// fmt.Printf("doSendJSONMetric() %v\r\n", metric)

	jsonValue, _ := json.Marshal(metric)
	// fmt.Printf("doSendJSONMetric(): %s\r\n", jsonValue)
	resp, err := http.Post(serverJSONPOSTUpdateURI, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		// fmt.Printf("doSendJSONMetric(): url=%v, error=%v\r\n", serverJSONPOSTUpdateURI, err)
		return
	}
	defer resp.Body.Close()
	// fmt.Printf("doSendJSONMetric(): url=%v, resp=%v\r\n", serverJSONPOSTUpdateURI, resp)
}
