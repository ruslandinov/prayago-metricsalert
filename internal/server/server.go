package server

import (
	"fmt"
	"net/http"
	"prayago-metricsalert/internal/memstorage"
	"strconv"
	"strings"
)

type Server struct {
}

func NewServer() *Server {
	mux := http.NewServeMux()
	// Go 1.22 rocks!
	mux.HandleFunc(`POST /update/{mtype}/{mname}/{mvalue}`, updateMetric)
	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}

	return &Server{}
}

func updateMetric(res http.ResponseWriter, req *http.Request) {
	// for local development and debug
	var body = "Header ===============\r\n"
	for k, v := range req.Header {
		body += fmt.Sprintf("%s: %v\r\n", k, v)
	}
	body += "Params ===============\r\n"
	body += fmt.Sprintf("mtype: %v\r\n", req.PathValue("mtype"))
	body += fmt.Sprintf("mname: %v\r\n", req.PathValue("mname"))
	body += fmt.Sprintf("mvalue: %v\r\n", req.PathValue("mvalue"))
	// res.Write([]byte(body))

	if req.Header.Get("Content-type") != "text/plain" {
		http.Error(res, "Wrong Content-type, only text/plain is allowed", http.StatusMethodNotAllowed)
		return
	}

	if mname := req.PathValue("mname"); mname == "" {
		http.Error(res, "Wrong metric name.", http.StatusNotFound)
		return
	}

	mvalueStr := req.PathValue("mvalue")
	if mvalueStr == "" {
		http.Error(res, "Wrong metric value. Only gauge or counter are supported", http.StatusBadRequest)
		return
	}

	mtype := req.PathValue("mtype")
	switch mtype {
	case memstorage.GaugeMetric:
		mvalue, err := strconv.ParseFloat(strings.TrimSpace(mvalueStr), 64)
		if err != nil {
			http.Error(res, fmt.Sprintf("Wrong metric value: %v\r\n", err), http.StatusBadRequest)
			return
		}
		body += fmt.Sprintf("Gauge metric value parsed successfully: %v\r\n", mvalue)

	case memstorage.CounterMetric:
		mvalue, err := strconv.ParseInt(strings.TrimSpace(mvalueStr), 10, 64)
		if err != nil {
			http.Error(res, fmt.Sprintf("Wrong metric value: %v\r\n", err), http.StatusBadRequest)
			return
		}
		body += fmt.Sprintf("Counter metric value parsed successfully: %v\r\n", mvalue)

	default:
		http.Error(res, "Wrong metric type. Only gauge or counter are supported", http.StatusBadRequest)
		return
	}

	// res.Write([]byte(body))
	res.Header().Set("content-type", "text/plain")
	res.WriteHeader(http.StatusOK)
}
