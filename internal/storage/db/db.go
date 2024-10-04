package db

import (
	"context"
	"database/sql"
	"fmt"
	"prayago-metricsalert/internal/logger"
	"prayago-metricsalert/internal/metrics"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Metric = metrics.Metric

type DBStorageConfig struct {
	ConnectionString string
}

type DBStorage struct {
	config DBStorageConfig
	db     *sql.DB
}

type DBStorager interface {
	Ping() bool
	UpdateMetric(metric Metric)
	UpdateBatch(metrics []Metric) error
}

func NewDBStorage(config DBStorageConfig) DBStorage {
	logger.LogSugar.Infof("DBStorage created, config: %v", config)

	db, err := sql.Open("pgx", config.ConnectionString)
	if err != nil {
		logger.LogSugar.Errorf("Ошибка подключения к БД: %v\r\n", err)
	}

	createMetricsTable(db)

	dbstorage := DBStorage{
		config,
		db,
	}

	return dbstorage
}

func (dbs DBStorage) Close() {
	dbs.db.Close()
}

func (dbs DBStorage) Ping() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := dbs.db.PingContext(ctx); err != nil {
		logger.LogSugar.Errorf("Ошибка пинга к БД: %v", err)
		return false
	}

	return true
}

func (dbs DBStorage) UpdateMetric(metric Metric) {
	logger.LogSugar.Infoln("UpdateMetric: metric=", metric)

	var result sql.Result
	var err error
	if metric.ISGauge() {
		result, err = doUpdateGauge(dbs.db, metric)
	} else {
		result, err = doUpdateCounter(dbs.db, metric)
	}

	if err != nil {
		logger.LogSugar.Errorln("UpdateMetric: err=", err)
	} else {
		logger.LogSugar.Infoln("UpdateMetric: result=", result)
	}
}

var updateGaugeMetricQuery = `
INSERT INTO metrics (name, type, gauge)
VALUES ($1, $2, $3)
ON CONFLICT(name)
DO UPDATE SET
  type = EXCLUDED.type,
  gauge = EXCLUDED.gauge;
`

func doUpdateGauge(db *sql.DB, metric Metric) (sql.Result, error) {
	return db.Exec(
		updateGaugeMetricQuery,
		metric.ID, metric.MType, metric.GetValueField(),
	)
}

var updateCounterMetricQuery = `
INSERT INTO metrics (name, type, counter)
VALUES ($1, $2, $3)
ON CONFLICT(name)
DO UPDATE SET
  type = EXCLUDED.type,
  counter = EXCLUDED.counter;
`

func doUpdateCounter(db *sql.DB, metric Metric) (sql.Result, error) {
	return db.Exec(
		updateCounterMetricQuery,
		metric.ID, metric.MType, metric.GetDeltaField(),
	)
}

func (dbs DBStorage) UpdateBatch(metrics []Metric) error {
	// logger.LogSugar.Infoln("UpdateBatch: metrics=", metrics)
	return BulkInsert(dbs.db, metrics)
}

func BulkInsert(db *sql.DB, metrics []Metric) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	for _, metric := range metrics {
		var query string
		var queryArgs []interface{}

		if metric.ISGauge() {
			query = updateGaugeMetricQuery
			queryArgs = append(queryArgs, metric.ID)
			queryArgs = append(queryArgs, metric.MType)
			queryArgs = append(queryArgs, metric.GetValueField())
		} else {
			query = updateCounterMetricQuery
			queryArgs = append(queryArgs, metric.ID)
			queryArgs = append(queryArgs, metric.MType)
			queryArgs = append(queryArgs, metric.GetDeltaField())
		}

		_, err := tx.ExecContext(context.TODO(), query, queryArgs...)

		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// А счастье было так возможно, пока не запустил тесты.
// Постгря не дает модифицировать запис дважды(второе срабатывае ON CONFLICT порождает ошибку)
// Оставлю пока на память этот код, много времени на него потратил
func _BulkInsert(db *sql.DB, metrics []Metric) error {
	var colCount = 4
	valueStrings := make([]string, 0, len(metrics))
	valueArgs := make([]interface{}, 0, len(metrics)*colCount)

	// формируем VALUES($n, $n+1, $n+2, etc)
	// и кладем значения соответстующих параметров в valueArgs
	// для дальнейшего использования в db.Exec
	i := 0
	for _, metric := range metrics {
		valueStr := "("
		for j := 1; j <= colCount; j++ {
			valueStr = valueStr + fmt.Sprintf("$%d", i*colCount+j)
			if j < colCount {
				valueStr += ", "
			}
		}
		valueStr = valueStr + ")"
		valueStrings = append(valueStrings, valueStr)

		appendMetricFields(&valueArgs, metric)
		i++
	}
	// logger.LogSugar.Infoln("BulkInsert: valueStrings=", strings.Join(valueStrings, ","))
	// logger.LogSugar.Infoln("BulkInsert: valueArgs=", valueArgs)
	query := fmt.Sprintf(
		"INSERT INTO metrics (name, type, gauge, counter) VALUES %s %s",
		strings.Join(valueStrings, ","),
		"ON CONFLICT(name) DO UPDATE SET type = EXCLUDED.type, gauge = EXCLUDED.gauge, counter = EXCLUDED.counter;",
	)
	// logger.LogSugar.Infoln("BulkInsert: query=", query)

	result, err := db.Exec(query, valueArgs...)
	if err != nil {
		logger.LogSugar.Errorln("UpdateMetric: err=", err)
	} else {
		logger.LogSugar.Infoln("UpdateMetric: result=", result)
	}

	return err
}

func appendMetricFields(args *[]interface{}, metric Metric) {
	*args = append(*args, metric.ID)
	*args = append(*args, metric.MType)
	if metric.ISGauge() {
		*args = append(*args, metric.GetValueField(), nil)
	} else {
		*args = append(*args, nil, metric.GetDeltaField())
	}
}
