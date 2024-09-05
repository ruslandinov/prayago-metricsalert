package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"prayago-metricsalert/internal/memstorage"

	"github.com/go-chi/chi/v5"
)

type (
	Metric = memstorage.Metric
	Server struct {
	}
)

func NewServer(ms memstorage.MemStorage, config ServerConfig) (*Server, error) {
	router := chi.NewRouter()
	router.Use(HTTPHandlerWithLogger)
	router.Get("/", gzipMiddleware(
		func(res http.ResponseWriter, req *http.Request) {
			getAllMetrics(ms, res, req)
		},
	))
	router.Get("/value/{mtype}/{mname}",
		func(res http.ResponseWriter, req *http.Request) {
			getMetric(ms, res, req)
		},
	)
	router.Post("/update/{mtype}/{mname}/{mvalue}",
		func(res http.ResponseWriter, req *http.Request) {
			updateMetric(ms, res, req)
		},
	)
	router.Post("/update/", gzipMiddleware(enforceContentTypeJSON(
		func(res http.ResponseWriter, req *http.Request) {
			updateMetricJSON(ms, res, req)
		},
	)))
	router.Post("/value/", gzipMiddleware(enforceContentTypeJSON(
		func(res http.ResponseWriter, req *http.Request) {
			getMetricJSON(ms, res, req)
		},
	)))

	err := http.ListenAndServe(config.ServerAddress, router)
	if err != nil {
		return nil, err
	}

	return &Server{}, nil
}

func getAllMetrics(ms memstorage.MemStorage, res http.ResponseWriter, _ *http.Request) {
	res.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := fmt.Sprintf("%s%s%s",
		`
<!DOCTYPE html>
<html lang="en">
    <head>
        <meta charset="utf-8">
        <title>Metrics</title>
    </head>
    <body>
`, ms.GetAllMetricsAsString(),
		`   </body>
</html>`)
	io.WriteString(res, html)
}

func getMetric(ms memstorage.MemStorage, res http.ResponseWriter, req *http.Request) {
	mName := chi.URLParam(req, "mname")

	value, err := ms.GetMetricValue(mName)
	if err != nil {
		http.Error(res, err.Error(), http.StatusNotFound)
		return
	}

	res.Header().Set("Content-type", "text/plain")
	res.WriteHeader(http.StatusOK)
	io.WriteString(res, fmt.Sprintf("%v", value))
}

func updateMetric(ms memstorage.MemStorage, res http.ResponseWriter, req *http.Request) {
	var mname string
	if mname = chi.URLParam(req, "mname"); mname == "" {
		http.Error(res, "empty metric name", http.StatusNotFound)
		return
	}

	mvalueStr := chi.URLParam(req, "mvalue")
	if mvalueStr == "" {
		http.Error(res, "empty metric value", http.StatusBadRequest)
		return
	}

	mtype := chi.URLParam(req, "mtype")
	if err := ms.UpdateMetricValue(mtype, mname, mvalueStr); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	}

	res.Header().Set("Content-type", "text/plain")
	res.WriteHeader(http.StatusOK)
}

func updateMetricJSON(ms memstorage.MemStorage, res http.ResponseWriter, req *http.Request) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	var metric Metric
	if err = json.Unmarshal(buf.Bytes(), &metric); err != nil {
		http.Error(res, "could not unmarshall JSON", http.StatusBadRequest)
		return
	}

	updatedMetric, err := ms.UpdateMetric(metric)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	}

	sendJSONedMetric(updatedMetric, res)
}

func getMetricJSON(ms memstorage.MemStorage, res http.ResponseWriter, req *http.Request) {
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

	updatedMetric, err := ms.GetMetric(metric.ID)
	if err != nil {
		http.Error(res, err.Error(), http.StatusNotFound)
		return
	}

	sendJSONedMetric(updatedMetric, res)
}

func sendJSONedMetric(metric *Metric, res http.ResponseWriter) {
	metricMarshalled, err := json.Marshal(&metric)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write(metricMarshalled)
}
