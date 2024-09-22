package server

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"prayago-metricsalert/internal/chimocker"
	"prayago-metricsalert/internal/metrics"

	"github.com/stretchr/testify/assert"
)

// т.к. сервер через DI получает ссылку на экземпляр Storage,
// то объявим здесь мок Storage, чтобы не использовать настоящий инстанс в тестах
type dummyStorage struct {
	getMetricValueImpl func(name string) (any, error)
}

func (store dummyStorage) GetAllMetricsAsString() string {
	return ""
}

func (store dummyStorage) GetMetricValue(name string) (any, error) {
	// если в моке есть реализация метода -- используем ее, иначе отадим пустое значение
	if store.getMetricValueImpl != nil {
		return store.getMetricValueImpl(name)
	}

	return nil, nil
}

func (store dummyStorage) GetMetric(name string) (*Metric, error) {
	return nil, nil
}

func (store dummyStorage) UpdateMetricValue(mType string, name string, value string) error {
	if name == "" {
		return errors.New("")
	}

	if mType != metrics.GaugeMetric && mType != metrics.CounterMetric {
		return errors.New("")
	}

	metric := metrics.NewMetric(name, mType)
	err := metric.UpdateValueStr(value)
	return err
}

func (store dummyStorage) UpdateMetric(metric Metric) (*Metric, error) {
	if metric.ID == "" {
		return nil, errors.New("")
	}

	if metric.MType != metrics.GaugeMetric && metric.MType != metrics.CounterMetric {
		return nil, errors.New("")
	}

	return nil, nil
}

func (store dummyStorage) SaveData() {
}

func (store dummyStorage) Ping() bool {
	return true
}

func TestUpdateMetric(t *testing.T) {
	// для теста этого хендлера нам сойдет максимально простой мок
	store := dummyStorage{}

	type want struct {
		code     int
		response string
	}
	tests := []struct {
		name   string
		mType  string
		mName  string
		mValue string
		want   want
	}{
		{
			name:   "Update gauge metric should be StatusOK",
			mType:  "gauge",
			mName:  "zzz",
			mValue: "1.555",
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name:   "Update counter metric should be StatusOK",
			mType:  "counter",
			mName:  "yyy",
			mValue: "42",
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name:   "Update unsupported type metric should return StatudBadRequest",
			mType:  "bool",
			mName:  "zzz",
			mValue: "true",
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name:   "Update bad named metric should return StatusNotFound",
			mType:  "counter",
			mName:  "",
			mValue: "45",
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name:   "Update wrong counter value should return StatusBadRequest",
			mType:  "counter",
			mName:  "counterMetric1",
			mValue: "xyz",
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name:   "Update wrong gauge value should return StatusBadRequest",
			mType:  "gauge",
			mName:  "gaugeMetric1",
			mValue: "abc",
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := fmt.Sprintf("/update/%s/%s/%s", test.mType, test.mName, test.mValue)
			// fmt.Printf("path: %s\r\n", path)
			request := httptest.NewRequest(http.MethodPost, path, nil)
			w := httptest.NewRecorder()

			// не можем мы просто так взять и протестировать переметризированные роутеры chi :(((
			// https://haykot.dev/blog/til-testing-parametrized-urls-with-chi-router/
			urlParams := chimocker.URLParams{"mtype": test.mType, "mname": test.mName, "mvalue": test.mValue}
			request = chimocker.WithURLParams(request, urlParams)
			updateMetric(store, w, request)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.code, res.StatusCode)
		})
	}
}

func TestGetMetric(t *testing.T) {
	// для теста этого хендлера нам нужен мок, способный возвращать данные из хранилища
	store := dummyStorage{
		getMetricValueImpl: func(name string) (any, error) {
			if name == "existingMetric" {
				return "42", nil
			}

			return nil, fmt.Errorf("value not found")
		},
	}

	type want struct {
		code     int
		response string
	}
	tests := []struct {
		name  string
		mType string
		mName string
		want  want
	}{
		{
			name:  "Get known metric should be StatusOK and return value",
			mType: "counter",
			mName: "existingMetric",
			want: want{
				code:     http.StatusOK,
				response: "42",
			},
		},
		{
			name:  "Get unknown metric should be StatusOK and return StatusNotFound",
			mType: "counter",
			mName: "unknownMetric",
			want: want{
				code:     http.StatusNotFound,
				response: "value not found\n",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := fmt.Sprintf("/value/%s/%s", test.mType, test.mName)
			// fmt.Printf("path: %s\r\n", path)
			request := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()

			// не можем мы просто так взять и протестировать переметризированные роутеры chi :(((
			// https://haykot.dev/blog/til-testing-parametrized-urls-with-chi-router/
			urlParams := chimocker.URLParams{"mtype": test.mType, "mname": test.mName}
			request = chimocker.WithURLParams(request, urlParams)

			getMetric(store, w, request)

			res := w.Result()
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			assert.Equal(t, test.want.code, res.StatusCode)
			if err == nil {
				assert.Equal(t, test.want.response, string(resBody))
			}
		})
	}
}
