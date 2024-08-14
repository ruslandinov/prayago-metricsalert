package memstorage

import (
	"fmt"
)

const (
	GaugeMetric   = "gauge"
	CounterMetric = "counter"
)

type MemStorage interface {
	StoreMetric(mType string, name string, value any)
	GetMetric(name string) (string, bool)
}

type memStorage struct {
	// storage["gauge"] -> map of gauge type metric values by name
	// storage["counter"] -> map of counter metric values by name
	storage map[string]interface{}
}

func NewMemStorage() MemStorage {
	var storage = make(map[string]interface{}, 0)

	return &memStorage{storage}
}

func (ms *memStorage) StoreMetric(mType string, name string, value any) {
	// fmt.Printf("StoreMetric: type=%v name=%v value=%v\r\n", mType, name, value)

	if mType == GaugeMetric {
		ms.storage[name] = value
		return
	}

	if oldValue, present := ms.storage[name]; present {
		ms.storage[name] = oldValue.(int64) + value.(int64)
	} else {
		ms.storage[name] = value.(int64)
	}
}

func (ms *memStorage) GetMetric(name string) (string, bool) {
	// fmt.Printf("GetMetric: name=%v\r\n", name)

	if value, present := ms.storage[name]; present {
		return fmt.Sprintf("%v", value), true
	}

	return "", false
}
