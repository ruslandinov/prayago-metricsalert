package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"prayago-metricsalert/internal/logger"
	"prayago-metricsalert/internal/storage"

	"github.com/go-chi/chi/v5"
)

type (
	Metric = storage.Metric
)

func GetRouter(store storage.Storager) http.Handler {
	router := chi.NewRouter()
	router.Use(HTTPHandlerWithLogger)
	router.Get("/", gzipMiddleware(
		func(res http.ResponseWriter, req *http.Request) {
			getAllMetrics(store, res, req)
		},
	))
	router.Get("/ping",
		func(res http.ResponseWriter, req *http.Request) {
			ping(store, res, req)
		},
	)
	router.Get("/value/{mtype}/{mname}",
		func(res http.ResponseWriter, req *http.Request) {
			getMetric(store, res, req)
		},
	)
	router.Post("/update/{mtype}/{mname}/{mvalue}",
		func(res http.ResponseWriter, req *http.Request) {
			updateMetric(store, res, req)
		},
	)
	router.Post("/update/", gzipMiddleware(enforceContentTypeJSON(
		func(res http.ResponseWriter, req *http.Request) {
			updateMetricJSON(store, res, req)
		},
	)))
	router.Post("/updates/", gzipMiddleware(enforceContentTypeJSON(
		func(res http.ResponseWriter, req *http.Request) {
			updatesBatch(store, res, req)
		},
	)))
	router.Post("/value/", gzipMiddleware(enforceContentTypeJSON(
		func(res http.ResponseWriter, req *http.Request) {
			getMetricJSON(store, res, req)
		},
	)))

	return router
}

func getAllMetrics(store storage.Storager, res http.ResponseWriter, _ *http.Request) {
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
`, store.GetAllMetricsAsString(),
		`   </body>
</html>`)
	io.WriteString(res, html)
}

func ping(store storage.Storager, res http.ResponseWriter, _ *http.Request) {
	if store.Ping() {
		res.WriteHeader(http.StatusOK)
		return
	}

	http.Error(res, "No database connection", http.StatusInternalServerError)
}

func getMetric(store storage.Storager, res http.ResponseWriter, req *http.Request) {
	mName := chi.URLParam(req, "mname")

	value, err := store.GetMetricValue(mName)
	if err != nil {
		http.Error(res, err.Error(), http.StatusNotFound)
		return
	}

	res.Header().Set("Content-type", "text/plain")
	res.WriteHeader(http.StatusOK)
	io.WriteString(res, fmt.Sprintf("%v", value))
}

func updateMetric(store storage.Storager, res http.ResponseWriter, req *http.Request) {
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
	if _, err := store.UpdateMetricValue(mtype, mname, mvalueStr); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	}

	res.Header().Set("Content-type", "text/plain")
	res.WriteHeader(http.StatusOK)
}

func updateMetricJSON(store storage.Storager, res http.ResponseWriter, req *http.Request) {
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

	updatedMetric, err := store.UpdateMetric(metric)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	}

	sendJSONedMetric(updatedMetric, res)
}

func updatesBatch(store storage.Storager, res http.ResponseWriter, req *http.Request) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		logger.LogSugar.Errorln("updatesBatch() err:", err)
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	var metrics = make([]Metric, 0)
	if err = json.Unmarshal(buf.Bytes(), &metrics); err != nil {
		logger.LogSugar.Errorln("updatesBatch() err:", err)
		http.Error(res, "could not unmarshall JSON", http.StatusBadRequest)
		return
	}

	// logger.LogSugar.Infoln("updatesBatch() metrics:", metrics)
	err = store.UpdateBatch(metrics)
	if err != nil {
		logger.LogSugar.Errorln("updatesBatch() err:", err)
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
}

func getMetricJSON(store storage.Storager, res http.ResponseWriter, req *http.Request) {
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

	updatedMetric, err := store.GetMetric(metric.ID)
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
