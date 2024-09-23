package db

import (
	"database/sql"
	"prayago-metricsalert/internal/logger"
)

func createMetricsTable(db *sql.DB) {
	// TODO: как передать имя таблицы параметром? $1 не работает, пришлось хардкодить
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS metrics (
			name VARCHAR(50) PRIMARY KEY,
			type VARCHAR(20) NOT NULL,
			gauge DOUBLE PRECISION,
			counter BIGINT
		);
	`)

	if err != nil {
		logger.LogSugar.Errorf("Ошибка создания таблицы: %v", err)
		panic(err)
	}
}
