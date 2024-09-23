package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"prayago-metricsalert/internal/logger"
	"prayago-metricsalert/internal/metrics"
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

type Metric = metrics.Metric

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
	logger.LogSugar.Infoln("Agent created")

	serverJSONPOSTUpdateURI = fmt.Sprintf("http://%s/update/", config.serverAddress)

	return &Agent{
		config:      config,
		metrics:     make(map[string]Metric),
		pollCount:   metrics.NewMetric("PollCount", metrics.CounterMetric),
		randomValue: metrics.NewMetric("RandomValue", metrics.GaugeMetric),
	}
}

func (agent *Agent) Run() {
	logger.LogSugar.Infoln("Agent started")
	go agent.startPolling()
	go agent.startSending()
	for {
		time.Sleep(1 * time.Second)
	}
}

func (agent *Agent) startPolling() {
	logger.LogSugar.Infoln("Agent started polling", agent.config.pollInterval)
	for {
		time.Sleep(agent.config.pollInterval)
		agent.updateMetrics()
	}
}

func (agent *Agent) startSending() {
	logger.LogSugar.Infoln("Agent started sending", agent.config.reportInterval)
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
			agent.metrics[mName] = Metric{ID: mName, MType: metrics.GaugeMetric, Value: &valueFloat64}
		}
	}

	*agent.pollCount.Delta++
	*agent.randomValue.Value = rand.Float64()
	// fmt.Printf("%v \r\n\r\n", agent.metrics)
}

func (agent *Agent) sendMetrics() {
	var url string
	for _, metric := range agent.metrics {
		url = fmt.Sprintf("http://%s/update/%s/%s/%v",
			agent.config.serverAddress,
			metric.MType, metric.ID, *metric.Value,
		)
		doSendMetric(url)
	}

	// pollCount
	url = fmt.Sprintf("http://%s/update/%s/%s/%v",
		agent.config.serverAddress,
		metrics.CounterMetric, pollCount, *agent.pollCount.Delta,
	)
	doSendMetric(url)
	// randomValue
	url = fmt.Sprintf("http://%s/update/%s/%s/%v",
		agent.config.serverAddress,
		metrics.GaugeMetric, randomValue, *agent.randomValue.Value,
	)
	doSendMetric(url)
}

func doSendMetric(url string) {
	resp, err := http.Post(url, "text/plain", nil)
	if err != nil {
		logger.LogSugar.Errorf("doSendMetric(): url=%v, error=%v\r\n", url, err)
		return
	}
	defer resp.Body.Close()
}

func (agent *Agent) sendJSONMetrics() {
	for _, metric := range agent.metrics {
		doSendJSONMetric(metric)
	}

	doSendJSONMetric(agent.pollCount)
	doSendJSONMetric(agent.randomValue)
}

func doSendJSONMetric(metric Metric) {
	jsonValue, _ := json.Marshal(metric)

	var gzippedBytes bytes.Buffer
	gzipper := gzip.NewWriter(&gzippedBytes)
	gzipOk := false
	if _, err := gzipper.Write(jsonValue); err == nil {
		if err := gzipper.Close(); err == nil {
			gzipOk = true
		}
	}

	var req *http.Request
	if gzipOk {
		req, _ = http.NewRequest("POST", serverJSONPOSTUpdateURI, &gzippedBytes)
		req.Header.Add("Content-Encoding", "gzip")
	} else {
		// resp, err := http.Post(serverJSONPOSTUpdateURI, "application/json", bytes.NewBuffer(jsonValue))
		req, _ = http.NewRequest("POST", serverJSONPOSTUpdateURI, bytes.NewBuffer(jsonValue))
	}

	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.LogSugar.Errorln("doSendJSONMetric", "error", err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		logger.LogSugar.Warnln("doSendJSONMetric",
			"status:", resp.StatusCode,
			"response:", string(bodyBytes),
		)
	}
	defer resp.Body.Close()
}
