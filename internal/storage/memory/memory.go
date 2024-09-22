package memory

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"prayago-metricsalert/internal/logger"
	"prayago-metricsalert/internal/metrics"
	"time"
)

type Metric = metrics.Metric

type MemStorageConfig struct {
	FPath         string
	StoreInterval time.Duration
	ShouldRestore bool
}

type MemStorage struct {
	config  MemStorageConfig
	storage map[string]Metric
}

func runStoreInteval(ms MemStorage) {
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

	memStorage := MemStorage{
		config:  config,
		storage: storage,
	}

	if config.ShouldRestore {
		memStorage.restoreData()
	}

	if config.StoreInterval > 0 {
		runStoreInteval(memStorage)
	}

	return memStorage
}

func (ms MemStorage) GetAllMetricsAsString() string {
	s := ""
	for _, metric := range ms.storage {
		s += fmt.Sprintf("%s|%s|%s|\r\n", metric.ID, metric.MType, metric.GetValueStr())
	}

	return s
}

func (ms MemStorage) GetMetricValue(name string) (any, error) {
	if metric, present := ms.storage[name]; present {
		return metric.GetValue(), nil
	}

	return nil, errors.New("metric not found")
}

func (ms MemStorage) GetMetric(name string) (*Metric, error) {
	if metric, present := ms.storage[name]; present {
		return &metric, nil
	}

	return nil, errors.New("metric not found")
}

func (ms MemStorage) UpdateMetricValue(mType string, name string, value string) error {
	if mType != metrics.GaugeMetric && mType != metrics.CounterMetric {
		return fmt.Errorf("unsupported metric type %s", mType)
	}

	metric, present := ms.storage[name]
	if !present {
		metric = metrics.NewMetric(name, mType)
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

func (ms MemStorage) UpdateMetric(metric Metric) (*Metric, error) {
	if metric.MType != metrics.GaugeMetric && metric.MType != metrics.CounterMetric {
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

	if err := oldMetric.UpdateValueNum(metric.GetValue()); err != nil {
		return nil, err
	}

	if ms.config.StoreInterval == 0 {
		ms.SaveData()
	}

	return &oldMetric, nil
}

func (ms MemStorage) SaveData() {
	logger.LogSugar.Infof("Memstorage saving, config %v", ms.config)
	logger.LogSugar.Infoln("Memstorage saving to file", ms.config.FPath)

	data, err := json.Marshal(ms.storage)
	if err != nil {
		logger.LogSugar.Errorf("error JSONing metrics: %v", err)
	}

	if err := os.WriteFile(ms.config.FPath, data, 0666); err != nil {
		logger.LogSugar.Errorf("error saving file: %v", err)
	}
}

func (ms MemStorage) restoreData() {
	logger.LogSugar.Infoln("Memstorage restoring data from file", ms.config.FPath)

	data, err := os.ReadFile(ms.config.FPath)
	if err != nil {
		logger.LogSugar.Errorf("error reading file: %v", err)
	}

	if err := json.Unmarshal(data, &ms.storage); err != nil {
		logger.LogSugar.Errorf("error parsing JSONing metrics: %v", err)
	}
}
