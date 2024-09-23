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
	dbstore  db.DBStorager
}

type Storager interface {
	GetAllMetricsAsString() string
	GetMetricValue(name string) (any, error)
	GetMetric(name string) (*Metric, error)
	UpdateMetricValue(mType string, name string, value string) (*Metric, error)
	UpdateMetric(metric Metric) (*Metric, error)
	UpdateBatch(metrics []Metric) error
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

	// не хочется раскидывать по коду проверки "если есть база, то пиши в нее"
	// поэтому если база не нужна, создаю "мок", который ничего не делает, и код чище
	var dbstore db.DBStorager
	if config.DBConnectionString != "" {
		dbsConfig := db.DBStorageConfig{
			ConnectionString: config.DBConnectionString,
		}
		dbstore = db.NewDBStorage(dbsConfig)
	} else {
		dbstore = db.NewDummyDB()
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

func (st Storage) UpdateMetricValue(mType string, name string, value string) (*Metric, error) {
	metric, err := st.memstore.UpdateMetricValue(mType, name, value)
	if err == nil {
		st.dbstore.UpdateMetric(*metric)
	}

	return metric, err
}

func (st Storage) UpdateMetric(metric Metric) (*Metric, error) {
	updatedMetric, err := st.memstore.UpdateMetric(metric)
	if err == nil {
		st.dbstore.UpdateMetric(*updatedMetric)
	}

	return updatedMetric, err
}

func (st Storage) UpdateBatch(metrics []Metric) error {
	return st.dbstore.UpdateBatch(metrics)
}

func (st Storage) SaveData() {
	st.memstore.SaveData()
}

func (st Storage) Ping() bool {
	return st.dbstore.Ping()
}
