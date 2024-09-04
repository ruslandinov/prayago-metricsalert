package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"prayago-metricsalert/internal/logger"
	"prayago-metricsalert/internal/memstorage"
	"prayago-metricsalert/internal/protocol"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

type (
	Metric = protocol.Metric
	Server struct {
	}
)

func NewServer(ms memstorage.MemStorage, config ServerConfig) *Server {
	router := chi.NewRouter()

	router.Get("/", logger.HTTPHandlerWithLogger(
		func(res http.ResponseWriter, req *http.Request) {
			getAllMetrics(ms, res, req)
		},
	))
	router.Get("/value/{mtype}/{mname}", logger.HTTPHandlerWithLogger(
		func(res http.ResponseWriter, req *http.Request) {
			getMetric(ms, res, req)
		},
	))
	router.Post("/update/", logger.HTTPHandlerWithLogger(
		func(res http.ResponseWriter, req *http.Request) {
			updateMetricJSON(ms, res, req)
		},
	))
	router.Post("/update/{mtype}/{mname}/{mvalue}", logger.HTTPHandlerWithLogger(
		func(res http.ResponseWriter, req *http.Request) {
			updateMetric(ms, res, req)
		},
	))

	err := http.ListenAndServe(config.serverAddress, router)
	if err != nil {
		panic(err)
	}

	return &Server{}
}

func getAllMetrics(ms memstorage.MemStorage, res http.ResponseWriter, _ *http.Request) {
	io.WriteString(res, ms.GetAllMetricsAsString())
}

func getMetric(ms memstorage.MemStorage, res http.ResponseWriter, req *http.Request) {
	// mType := chi.URLParam(req, "mtype")
	mName := chi.URLParam(req, "mname")

	if value, present := ms.GetMetric(mName); present {
		io.WriteString(res, fmt.Sprintf("%v", value))
		return
	}

	http.Error(res, "Wrong metric name.", http.StatusNotFound)
}

func updateMetric(ms memstorage.MemStorage, res http.ResponseWriter, req *http.Request) {
	var mname string
	if mname = chi.URLParam(req, "mname"); mname == "" {
		http.Error(res, "Wrong metric name.", http.StatusNotFound)
		return
	}

	mvalueStr := chi.URLParam(req, "mvalue")
	if mvalueStr == "" {
		http.Error(res, "Empty metric value.", http.StatusBadRequest)
		return
	}

	mtype := chi.URLParam(req, "mtype")
	if _, err := storeMetricValue(ms, mtype, mname, mvalueStr); err != nil {
		http.Error(res, fmt.Sprintf("Wrong metric value: %v\r\n", err), http.StatusBadRequest)
	}

	res.Header().Set("content-type", "text/plain")
	res.WriteHeader(http.StatusOK)
}

func updateMetricJSON(ms memstorage.MemStorage, res http.ResponseWriter, req *http.Request) {
	// TODO: extract content type check to middleware
	ctHeader := req.Header.Get("Content-Type")
	if ctHeader != "" {
		contentType := strings.ToLower(strings.TrimSpace(strings.Split(ctHeader, ";")[0]))
		if contentType != "application/json" {
			http.Error(res, "Wrong content type", http.StatusBadRequest)
		}
	}

	var buf bytes.Buffer
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	var metric Metric
	if err = json.Unmarshal(buf.Bytes(), &metric); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	// fmt.Printf("updateMetricJSON: metric=%v\r\n", metric)

	// TODO: should we validate metric values depending on metric type?
	var mValue any
	if metric.MType == memstorage.GaugeMetric {
		mValue = metric.Value
	} else {
		mValue = metric.Delta
	}

	var newValue any
	if newValue, err = storeMetricValue(ms, metric.MType, metric.ID, mValue); err != nil {
		http.Error(res, fmt.Sprintf("Wrong metric value: %v\r\n", err), http.StatusBadRequest)
	}

	if metric.MType == memstorage.GaugeMetric {
		metric.Value = newValue.(float64)
	} else {
		metric.Delta = newValue.(int64)
	}

	metricMarshalled, err := json.Marshal(metric)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write(metricMarshalled)
}

func storeMetricValue(ms memstorage.MemStorage, mType string, mName string, mValue any) (any, error) {
	typedValue, err := typeMetricValue(mType, mValue)
	if err != nil {
		logger.LogSugar.Infoln("storeMetricValue", "error", err)
		return nil, err
	}

	return ms.StoreMetric(mType, mName, typedValue), nil
}

// used to handle text/plain and application/json ways of passing metrics value
// text/plain -- metric value is passed as string
// application/json -- after unmarshalling metric value MUST be either float64 or int64
func typeMetricValue(mType string, value any) (any, error) {
	// fmt.Printf("mType, value, value type: %v, %v, %v \r\n", mType, value, reflect.TypeOf(value))
	switch assertedTypeValue := value.(type) {
	case float64:
		return value, nil
	case int64:
		return value, nil
	case string:
		if mType == memstorage.GaugeMetric {
			typedValue, err := strconv.ParseFloat(strings.TrimSpace(assertedTypeValue), 64)
			if err != nil {
				return nil, err
			}
			return typedValue, nil
		}

		typedValue, err := strconv.ParseInt(strings.TrimSpace(assertedTypeValue), 10, 64)
		if err != nil {
			return nil, err
		}
		return typedValue, nil
	default:
		return nil, errors.New("unsupported metric type")
	}
}
