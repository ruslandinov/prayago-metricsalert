package metrics

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	GaugeMetric   = "gauge"
	CounterMetric = "counter"
)

type (
	Metric struct {
		ID    string   `json:"id"`              // имя метрики
		MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
		Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
		Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	}
)

func (m *Metric) UnmarshalJSON(data []byte) error {
	// fmt.Printf("UnmarshalJSON start: %s\r\n", string(data[:]))

	type MyMetricAlias Metric
	aliasValue := &struct {
		*MyMetricAlias
		Value json.RawMessage `json:"value"`
	}{
		MyMetricAlias: (*MyMetricAlias)(m),
	}

	if err := json.Unmarshal(data, aliasValue); err != nil {
		fmt.Printf("UnmarshalJSON error %v\r\n", err)
		return err
	}

	// json.Marshall превращает 0.0(float64) в 0(int64)
	// что приводит к тому, что при json.Unmarshall метрика типа gauge
	// получает значение типа int64 и сервер в скором времени сыпется
	// поэтому кастомный UnmarshalJSON
	if m.MType == GaugeMetric && aliasValue.Value != nil {
		var valueFloat64 float64
		json.Unmarshal(aliasValue.Value, &valueFloat64)
		m.Value = &valueFloat64
	} else {
		json.Unmarshal(aliasValue.Value, &m.Value)
	}

	// if m.Value != nil {
	// 	fmt.Printf("UnmarshalJSON end: %v:%v\r\n", *m.Value, reflect.TypeOf(*m.Value))
	// }

	// if m.Delta != nil {
	// 	fmt.Printf("UnmarshalJSON end: %v:%v\r\n", *m.Delta, reflect.TypeOf(*m.Delta))
	// }

	return nil
}

func NewMetric(ID string, MType string) Metric {
	var zeroInt int64 = 0
	var zeroFloat float64 = 0

	if MType == GaugeMetric {
		return Metric{
			ID:    ID,
			MType: MType,
			Value: &zeroFloat,
		}
	}

	return Metric{
		ID:    ID,
		MType: MType,
		Delta: &zeroInt,
	}
}

func (m Metric) GetDeltaField() int64 {
	return *m.Delta
}

func (m Metric) GetValueField() float64 {
	return *m.Value
}

func (m Metric) ISGauge() bool {
	return m.MType == GaugeMetric
}

func (m Metric) GetValue() any {
	if m.MType == GaugeMetric {
		return *m.Value
	}

	return *m.Delta
}

func (m Metric) GetValueStr() string {
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
		*m.Value = value.(float64)
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
