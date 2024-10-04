package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"prayago-metricsalert/internal/logger"
	"prayago-metricsalert/internal/metrics"
	"reflect"
	"runtime"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
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
	client      *resty.Client
}

const pollCount = "PollCount"
const randomValue = "RandomValue"

var updateMetricURI string
var batchUpdateMetricsURI string

func NewAgent(config AgentConfig) *Agent {
	logger.LogSugar.Infoln("Agent created")

	updateMetricURI = fmt.Sprintf("http://%s/update/", config.serverAddress)
	batchUpdateMetricsURI = fmt.Sprintf("http://%s/updates/", config.serverAddress)

	// под 13ый инкремент
	client := resty.New()
	client.
		SetRetryCount(3).
		SetRetryMaxWaitTime(15 * time.Second).
		SetRetryAfter(
			func(client *resty.Client, resp *resty.Response) (time.Duration, error) {
				if resp.Request.Attempt > 4 {
					return 0, errors.New("quota exceeded")
				}
				// 3 попытки, через 1-3-5 секунд, арифметика второй класс
				return time.Duration(resp.Request.Attempt*2-1) * time.Second, nil
			},
		).
		AddRetryCondition(
			func(r *resty.Response, err error) bool {
				return err != nil || r.StatusCode() == http.StatusTooManyRequests
			},
		)

	return &Agent{
		config:      config,
		metrics:     make(map[string]Metric),
		pollCount:   metrics.NewMetric("PollCount", metrics.CounterMetric),
		randomValue: metrics.NewMetric("RandomValue", metrics.GaugeMetric),
		client:      client,
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
		agent.sendMetricsBatch()
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
		doPostMetric(agent.client, url)
	}

	// pollCount
	url = fmt.Sprintf("http://%s/update/%s/%s/%v",
		agent.config.serverAddress,
		metrics.CounterMetric, pollCount, *agent.pollCount.Delta,
	)
	doPostMetric(agent.client, url)
	// randomValue
	url = fmt.Sprintf("http://%s/update/%s/%s/%v",
		agent.config.serverAddress,
		metrics.GaugeMetric, randomValue, *agent.randomValue.Value,
	)
	doPostMetric(agent.client, url)
}

func doPostMetric(client *resty.Client, url string) {
	_, err := client.R().SetHeader("Content-Type", "text/plain").Post(url)
	if err != nil {
		logger.LogSugar.Errorf("doPostMetric(): url=%v, error=%v", url, err)
		return
	}
}

func (agent *Agent) sendJSONMetrics() {
	for _, metric := range agent.metrics {
		doSendJSONMetric(agent.client, metric)
	}

	doSendJSONMetric(agent.client, agent.pollCount)
	doSendJSONMetric(agent.client, agent.randomValue)
}

func doSendJSONMetric(client *resty.Client, metric Metric) {
	jsonValue, err := json.Marshal(metric)
	if err != nil {
		logger.LogSugar.Errorln("doSendJSONMetric", err)
		return
	}

	doPostJSON(*client, updateMetricURI, jsonValue)
}

func (agent *Agent) sendMetricsBatch() {
	var metricsSlice = make([]Metric, 0)
	for _, metric := range agent.metrics {
		metricsSlice = append(metricsSlice, metric)
	}
	metricsSlice = append(metricsSlice, agent.pollCount, agent.randomValue)

	jsonValue, err := json.Marshal(metricsSlice)
	if err != nil {
		logger.LogSugar.Errorln("sendMetricsBatch", err)
		return
	}

	// logger.LogSugar.Infoln("sendMetricsBatch: metrics=", string(jsonValue))
	doPostJSON(*agent.client, batchUpdateMetricsURI, jsonValue)
}

func doPostJSON(client resty.Client, url string, jsonValue []byte) {
	var gzippedBytes bytes.Buffer
	gzipper := gzip.NewWriter(&gzippedBytes)
	gzipOk := false
	if _, err := gzipper.Write(jsonValue); err == nil {
		if err := gzipper.Close(); err == nil {
			gzipOk = true
		}
	}

	var resp *resty.Response
	var err error
	req := client.R()
	req.SetHeader("Content-Type", "application/json")
	if gzipOk {
		req.SetHeader("Content-Encoding", "gzip")
		req.SetBody(&gzippedBytes)
		resp, err = req.Post(url)
	} else {
		req.SetBody(bytes.NewBuffer(jsonValue))
		resp, err = req.Post(url)
	}

	if err != nil {
		logger.LogSugar.Errorln("doPostJSON", "error", err)
		return
	} else {
		logger.LogSugar.Infoln("doPostJSON", "response:", string(resp.Body()))
	}

	if resp.StatusCode() != http.StatusOK {
		logger.LogSugar.Infoln("doPostJSON", "status:", resp.StatusCode())
	}
}
