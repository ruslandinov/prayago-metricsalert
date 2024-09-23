package db

import (
	"context"
	"database/sql"
	"prayago-metricsalert/internal/logger"
	"prayago-metricsalert/internal/metrics"
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
	logger.LogSugar.Infof("UpdateMetric: metric=", metric)

	var result sql.Result
	var err error
	if metric.ISGauge() {
		result, err = doUpdateGauge(dbs.db, metric)
	} else {
		result, err = doUpdateCounter(dbs.db, metric)
	}

	if err != nil {
		logger.LogSugar.Errorf("UpdateMetric: err=", err)
	} else {
		logger.LogSugar.Infof("UpdateMetric: result=", result)
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
