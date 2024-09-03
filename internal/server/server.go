package server

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"prayago-metricsalert/internal/logger"
	"prayago-metricsalert/internal/memstorage"
	"strconv"
	"strings"
)

type Server struct {
}

func NewServer(ms memstorage.MemStorage, config ServerConfig) *Server {
	router := chi.NewRouter()
	router.Get("/",  logger.HttpHandlerWithLogger(
		func(res http.ResponseWriter, req *http.Request) {
			getAllMetrics(ms, res, req)
		},
	))
	router.Get("/value/{mtype}/{mname}",  logger.HttpHandlerWithLogger(
		func(res http.ResponseWriter, req *http.Request) {
			getMetric(ms, res, req)
		},
	))
	router.Post("/update/{mtype}/{mname}/{mvalue}", logger.HttpHandlerWithLogger(
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
	switch mtype {
	case memstorage.GaugeMetric:
		mvalue, err := strconv.ParseFloat(strings.TrimSpace(mvalueStr), 64)
		if err != nil {
			http.Error(res, fmt.Sprintf("Wrong metric value: %v\r\n", err), http.StatusBadRequest)
			return
		}
		ms.StoreMetric(mtype, mname, mvalue)

	case memstorage.CounterMetric:
		mvalue, err := strconv.ParseInt(strings.TrimSpace(mvalueStr), 10, 64)
		if err != nil {
			http.Error(res, fmt.Sprintf("Wrong metric value: %v\r\n", err), http.StatusBadRequest)
			return
		}
		ms.StoreMetric(mtype, mname, mvalue)

	default:
		http.Error(res, "Wrong metric type. Only gauge or counter are supported", http.StatusBadRequest)
		return
	}

	res.Header().Set("content-type", "text/plain")
	res.WriteHeader(http.StatusOK)
}
