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
	GetMetric(mType string, name string) (string, bool)
}

type memStorage struct {
	// storage["gauge"] -> map of gauge type metric values by name
	// storage["counter"] -> map of counter metric values by name
	storage map[string]map[string]interface{}
}

func NewMemStorage() MemStorage {
	var storage = make(map[string]map[string]interface{}, 0)

	return &memStorage{storage}
}

func (ms *memStorage) StoreMetric(mType string, name string, value any) {
	fmt.Printf("StoreMetric: type=%v name=%v value=%v\r\n", mType, name, value)

	if _, present := ms.storage[mType]; !present {
		ms.storage[mType] = make(map[string]interface{})
	}
	ms.storage[mType][name] = value
}

func (ms *memStorage) GetMetric(mType string, name string) (string, bool) {
	fmt.Printf("GetMetric: type=%v name=%v\r\n", mType, name)

	if typeList, present := ms.storage[mType]; present {
		if value, present := typeList[name]; present {
			return fmt.Sprintf("%v", value), true
		}
	}

	return "", false
}
