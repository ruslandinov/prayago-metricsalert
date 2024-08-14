package server

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"prayago-metricsalert/internal/memstorage"
	"strconv"
	"strings"
)

type Server struct {
}

func NewServer(ms memstorage.MemStorage) *Server {
	router := chi.NewRouter()
	router.Post("/update/{mtype}/{mname}/{mvalue}", func(res http.ResponseWriter, req *http.Request) {
		updateMetric(ms, res, req)
	})

	err := http.ListenAndServe(`:8080`, router)
	if err != nil {
		panic(err)
	}

	return &Server{}
}

func updateMetric(ms memstorage.MemStorage, res http.ResponseWriter, req *http.Request) {
	// for local development and debug
	var body = "Header ===============\r\n"
	for k, v := range req.Header {
		body += fmt.Sprintf("%s: %v\r\n", k, v)
	}
	body += "Params ===============\r\n"
	body += fmt.Sprintf("mtype: %v\r\n", chi.URLParam(req, "mtype"))
	body += fmt.Sprintf("mname: %v\r\n", chi.URLParam(req, "mname"))
	body += fmt.Sprintf("mvalue: %v\r\n", chi.URLParam(req, "mvalue"))
	// res.Write([]byte(body))

	if req.Header.Get("Content-type") != "text/plain" {
		http.Error(res, "Wrong Content-type, only text/plain is allowed", http.StatusMethodNotAllowed)
		return
	}

	var mname string
	if mname = chi.URLParam(req, "mname"); mname == "" {
		http.Error(res, "Wrong metric name.", http.StatusNotFound)
		return
	}

	mvalueStr := chi.URLParam(req, "mvalue")
	if mvalueStr == "" {
		http.Error(res, "Wrong metric value. Only gauge or counter are supported", http.StatusBadRequest)
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
		body += fmt.Sprintf("Gauge metric value parsed successfully: %v\r\n", mvalue)

	case memstorage.CounterMetric:
		mvalue, err := strconv.ParseInt(strings.TrimSpace(mvalueStr), 10, 64)
		if err != nil {
			http.Error(res, fmt.Sprintf("Wrong metric value: %v\r\n", err), http.StatusBadRequest)
			return
		}
		ms.StoreMetric(mtype, mname, mvalue)
		body += fmt.Sprintf("Counter metric value parsed successfully: %v\r\n", mvalue)

	default:
		http.Error(res, "Wrong metric type. Only gauge or counter are supported", http.StatusBadRequest)
		return
	}
	// res.Write([]byte(body))
	res.Header().Set("content-type", "text/plain")
	res.WriteHeader(http.StatusOK)
}
