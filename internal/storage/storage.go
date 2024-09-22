package storage

import (
	"prayago-metricsalert/internal/metrics"
	"prayago-metricsalert/internal/storage/db"
	"prayago-metricsalert/internal/storage/memory"
	"time"
)

type Metric = metrics.Metric

type StorageConfig struct {
	FPath              string
	StoreInterval      time.Duration
	ShouldRestore      bool
	DBConnectionString string
}

type Storage struct {
	config   StorageConfig
	memstore memory.MemStorage
	dbstore  db.DBStorage
}

type Storager interface {
	GetAllMetricsAsString() string
	GetMetricValue(name string) (any, error)
	GetMetric(name string) (*Metric, error)
	UpdateMetricValue(mType string, name string, value string) error
	UpdateMetric(metric Metric) (*Metric, error)
	SaveData()
	Ping() bool
}

func NewStorage(config StorageConfig) Storage {
	msConfig := memory.MemStorageConfig{
		FPath:         config.FPath,
		StoreInterval: config.StoreInterval,
		ShouldRestore: config.ShouldRestore,
	}
	memstore := memory.NewMemStorage(msConfig)

	var dbstore = db.DBStorage{}
	if config.DBConnectionString != "" {
		dbsConfig := db.DBStorageConfig{
			ConnectionString: config.DBConnectionString,
		}
		dbstore = db.NewDBStorage(dbsConfig)
	}

	storage := Storage{
		config:   config,
		memstore: memstore,
		dbstore:  dbstore,
	}

	return storage
}

func (st Storage) GetAllMetricsAsString() string {
	return st.memstore.GetAllMetricsAsString()
}

func (st Storage) GetMetricValue(name string) (any, error) {
	return st.memstore.GetMetricValue(name)
}

func (st Storage) GetMetric(name string) (*Metric, error) {
	return st.memstore.GetMetric(name)
}

func (st Storage) UpdateMetricValue(mType string, name string, value string) error {
	return st.memstore.UpdateMetricValue(mType, name, value)
}

func (st Storage) UpdateMetric(metric Metric) (*Metric, error) {
	return st.memstore.UpdateMetric(metric)
}

func (st Storage) SaveData() {
	st.memstore.SaveData()
}

func (st Storage) Ping() bool {
	// как проверять на nil?
	if st.dbstore != (db.DBStorage{}) {
		return st.dbstore.Ping()
	}

	return false
}
