package memstorage

import (
	"fmt"
)

const (
	GaugeMetric   = "gauge"
	CounterMetric = "counter"
)

type MemStorage interface {
	StoreMetric(mType string, name string, value any) any
	GetMetric(name string) (any, bool)
	GetAllMetricsAsString() string
}

type memStorage struct {
	storage map[string]interface{}
}

func NewMemStorage() MemStorage {
	var storage = make(map[string]interface{}, 0)

	return &memStorage{storage}
}

func (ms *memStorage) StoreMetric(mType string, name string, value any) any {
	// fmt.Printf("StoreMetric: type=%v name=%v value=%v\r\n", mType, name, value)

	if mType == GaugeMetric {
		ms.storage[name] = value
	} else {
		if oldValue, present := ms.storage[name]; present {
			ms.storage[name] = oldValue.(int64) + value.(int64)
		} else {
			ms.storage[name] = value.(int64)
		}
	}

	return ms.storage[name]
}

func (ms *memStorage) GetMetric(name string) (any, bool) {
	// fmt.Printf("GetMetric: name=%v\r\n", name)

	if value, present := ms.storage[name]; present {
		return value, true
	}

	return nil, false
}

func (ms *memStorage) GetAllMetricsAsString() string {
	s := ""
	for mName, mValue := range ms.storage {
		s += fmt.Sprintf("%s=%v\r\n", mName, mValue)
	}

	return s
}
