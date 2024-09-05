package memstorage

import (
	"encoding/json"
	"errors"
	"fmt"
	"prayago-metricsalert/internal/logger"
)

const (
	GaugeMetric   = "gauge"
	CounterMetric = "counter"
)

type MemStorage interface {
	GetAllMetricsAsString() string
	GetMetricValue(name string) (any, error)
	GetMetric(name string) (*Metric, error)
	UpdateMetricValue(mType string, name string, value string) error
	UpdateMetric(metric Metric) (*Metric, error)
}

type MemStorageConfig struct {
	FPath         string
	StoreInterval time.Duration
	ShouldRestore bool
}

type memStorage struct {
	storage map[string]Metric
}

func NewMemStorage(config MemStorageConfig) MemStorage {
	var storage = make(map[string]Metric, 0)

	logger.LogSugar.Infof("MemStorage created, config: %v", config)

	memStorage := &memStorage{
		storage: storage,
	}

	return memStorage
}

func (ms *memStorage) GetAllMetricsAsString() string {
	s := ""
	for _, metric := range ms.storage {
		s += fmt.Sprintf("%s|%s|%s|\r\n", metric.ID, metric.MType, metric.getValueStr())
	}

	return s
}

func (ms *memStorage) GetMetricValue(name string) (any, error) {
	if metric, present := ms.storage[name]; present {
		return metric.getValue(), nil
	}

	return nil, errors.New("metric not found")
}

func (ms *memStorage) GetMetric(name string) (*Metric, error) {
	if metric, present := ms.storage[name]; present {
		return &metric, nil
	}

	return nil, errors.New("metric not found")
}

func (ms *memStorage) UpdateMetricValue(mType string, name string, value string) error {
	if mType != GaugeMetric && mType != CounterMetric {
		return fmt.Errorf("unsupported metric type %s", mType)
	}

	metric, present := ms.storage[name]
	if !present {
		metric = NewMetric(name, mType)
		ms.storage[name] = metric
	}

	if err := metric.UpdateValueStr(value); err != nil {
		return err
	}

	return nil
}

func (ms *memStorage) UpdateMetric(metric Metric) (*Metric, error) {
	if metric.MType != GaugeMetric && metric.MType != CounterMetric {
		return nil, fmt.Errorf("unsupported metric type %s", metric.MType)
	}

	if metric.ID == "" {
		return nil, fmt.Errorf("metric name is empty")
	}

	oldMetric, present := ms.storage[metric.ID]
	if !present {
		ms.storage[metric.ID] = metric
		return &metric, nil
	}

	if err := oldMetric.UpdateValueNum(metric.getValue()); err != nil {
		return nil, err
	}

	return &oldMetric, nil
}
