package memstorage

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type (
	Metric struct {
		ID    string   `json:"id"`              // имя метрики
		MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
		Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
		Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	}
)

func NewMetric(ID string, MType string) Metric {
	var zeroInt int64 = 0
	var zeroFloat float64 = 0

	return Metric{
		ID:    ID,
		MType: MType,
		Delta: &zeroInt,
		Value: &zeroFloat,
	}
}

func (m Metric) getValue() any {
	if m.MType == GaugeMetric {
		return *m.Value
	}

	return *m.Delta
}

func (m Metric) getValueStr() string {
	if m.MType == GaugeMetric {
		return strconv.FormatFloat(*m.Value, 'f', -1, 64)
	}

	return strconv.FormatInt(*m.Delta, 10)
}

func (m Metric) UpdateValueStr(value string) error {
	var typedValue any
	var err error
	if m.MType == GaugeMetric {
		typedValue, err = strconv.ParseFloat(strings.TrimSpace(value), 64)
	} else {
		typedValue, err = strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	}

	if err != nil {
		return err
	}

	m.doUpdateWithFloatOrIntValue(typedValue)
	return nil
}

func (m Metric) doUpdateWithFloatOrIntValue(value any) {
	if m.MType == GaugeMetric {
		*m.Value += value.(float64)
		return
	}

	*m.Delta += value.(int64)
}

func (m Metric) UpdateValueNum(value any) error {
	if m.MType == "" {
		return errors.New("missing type")
	}

	if m.MType != GaugeMetric && m.MType != CounterMetric {
		return fmt.Errorf("wrong metric type %s", m.MType)
	}

	switch assertedTypeValue := value.(type) {
	case float64:
		if m.MType != GaugeMetric {
			return fmt.Errorf("metric type and value mismatch %v:%v", m.MType, assertedTypeValue)
		}
	case int64:
		if m.MType != CounterMetric {
			return fmt.Errorf("metric type and value mismatch %v:%v", m.MType, assertedTypeValue)
		}
	case string:
	default:
		return fmt.Errorf("wrong value type %s", assertedTypeValue)
	}

	m.doUpdateWithFloatOrIntValue(value)
	return nil
}
