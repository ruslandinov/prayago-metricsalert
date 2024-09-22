package db

import (
	"database/sql"
	"fmt"
)

func createMetricsTable(db *sql.DB) {
	// TODO: как передать имя таблицы параметром? $1 не работает, пришлось хардкодить
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS metrics (
			name VARCHAR(50) PRIMARY KEY,
			type VARCHAR(20) NOT NULL,
			value DOUBLE PRECISION,
			counter BIGINT
		)
	`)

	if err != nil {
		fmt.Printf("Ошибка создания таблицы: %v\r\n", err)
		panic(err)
	}
}
