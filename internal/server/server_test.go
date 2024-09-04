package server

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"prayago-metricsalert/internal/chimocker"
)

// т.к. сервер через DI получает ссылку на экземпляр MemoryStorage,
// то объявим здесь мок MemoryStorage, чтобы не использовать настоящий инстанс в тестах
type dummyMemStorage struct {
	getMetricImpl             func(name string) (any, bool)
	getAllMetricsAsStringImpl func() string
}

func (ms dummyMemStorage) StoreMetric(mType string, name string, value any) any {
	return nil
}

func (ms dummyMemStorage) GetMetric(name string) (any, bool) {
	// если в моке есть реализация метода -- используем ее, иначе отадим пустое значение
	if ms.getMetricImpl != nil {
		return ms.getMetricImpl(name)
	}

	return nil, false
}

func (ms dummyMemStorage) GetAllMetricsAsString() string {
	// если в моке есть реализация метода -- используем ее, иначе отадим пустую сроку
	if ms.getAllMetricsAsStringImpl != nil {
		return ms.getAllMetricsAsStringImpl()
	}

	return ""
}

func TestUpdateMetric(t *testing.T) {
	// для теста этого хендлера нам сойдет максимально простой мок
	ms := dummyMemStorage{}

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
			updateMetric(ms, w, request)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.code, res.StatusCode)
		})
	}
}

func TestGetMetric(t *testing.T) {
	// для теста этого хендлера нам нужен мок, способный возвращать данные из хранилища
	ms := dummyMemStorage{
		getMetricImpl: func(name string) (any, bool) {
			if name == "existingMetric" {
				return "42", true
			}

			return nil, false
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
				response: "Wrong metric name.\n",
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

			getMetric(ms, w, request)

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
