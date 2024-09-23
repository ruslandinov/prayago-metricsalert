package db

type DummyDB struct{}

func NewDummyDB() DummyDB {
	return DummyDB{}
}

func (db DummyDB) Ping() bool {
	return false
}

func (db DummyDB) UpdateMetric(metric Metric) {
}

func (db DummyDB) UpdateBatch(metrics []Metric) error {
	return nil
}
