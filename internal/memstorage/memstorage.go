package memstorage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"prayago-metricsalert/internal/logger"
	"time"
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
	SaveData()
}

type MemStorageConfig struct {
	FPath         string
	StoreInterval time.Duration
	ShouldRestore bool
}

type memStorage struct {
	config  MemStorageConfig
	storage map[string]Metric
}

func runStoreInteval(ms memStorage) {
	time.AfterFunc(time.Millisecond*300, func() {
		for {
			time.Sleep(ms.config.StoreInterval)
			go ms.SaveData()
		}
	})
}

func NewMemStorage(config MemStorageConfig) MemStorage {
	var storage = make(map[string]Metric, 0)

	logger.LogSugar.Infof("MemStorage created, config: %v", config)

	memStorage := &memStorage{
		config:  config,
		storage: storage,
	}

	if config.ShouldRestore {
		memStorage.restoreData()
	}

	if config.StoreInterval > 0 {
		runStoreInteval(*memStorage)
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

	if ms.config.StoreInterval == 0 {
		ms.SaveData()
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

	if ms.config.StoreInterval == 0 {
		ms.SaveData()
	}

	return &oldMetric, nil
}

func (ms *memStorage) SaveData() {
	logger.LogSugar.Infoln("Memstorage saving to file", ms.config.FPath)

	data, err := json.Marshal(ms.storage)
	if err != nil {
		logger.LogSugar.Errorf("error JSONing metrics: %v", err)
	}

	if err := os.WriteFile(ms.config.FPath, data, 0666); err != nil {
		logger.LogSugar.Errorf("error saving file: %v", err)
	}
}

func (ms *memStorage) restoreData() {
	logger.LogSugar.Infoln("Memstorage restoring data from file", ms.config.FPath)

	data, err := os.ReadFile(ms.config.FPath)
	if err != nil {
		logger.LogSugar.Errorf("error reading file: %v", err)
	}

	if err := json.Unmarshal(data, &ms.storage); err != nil {
		logger.LogSugar.Errorf("error parsing JSONing metrics: %v", err)
	}
}
