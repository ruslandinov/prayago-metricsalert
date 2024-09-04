package server

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"prayago-metricsalert/internal/logger"
	"prayago-metricsalert/internal/memstorage"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

type Server struct {
}

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
		io.WriteString(res, value)
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

func storeMetricValue(ms memstorage.MemStorage, mType string, mName string, mValue string) (bool, error) {
	switch mType {
	case memstorage.GaugeMetric:
		mvalue, err := strconv.ParseFloat(strings.TrimSpace(mValue), 64)
		if err != nil {
			// TODO: use logger here
			fmt.Printf("Wrong gauge metric value: %v\r\n", err)
			return false, err
		}
		ms.StoreMetric(mType, mName, mvalue)

	case memstorage.CounterMetric:
		mvalue, err := strconv.ParseInt(strings.TrimSpace(mValue), 10, 64)
		if err != nil {
			// TODO: use logger here
			fmt.Printf("Wrong counter metric value: %v\r\n", err)
			return false, err
		}
		ms.StoreMetric(mType, mName, mvalue)

	default:
		return false, errors.New("wrong metric type, only gauge or counter are supported")
	}

	return true, nil
}
