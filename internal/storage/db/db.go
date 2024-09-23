package db

import (
	"context"
	"database/sql"
	"fmt"
	"prayago-metricsalert/internal/logger"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBStorageConfig struct {
	ConnectionString string
}

type DBStorage struct {
	config DBStorageConfig
	db     *sql.DB
}

func NewDBStorage(config DBStorageConfig) DBStorage {
	logger.LogSugar.Infof("DBStorage created, config: %v", config)

	db, err := sql.Open("pgx", config.ConnectionString)
	if err != nil {
		fmt.Printf("Ошибка подключения к БД: %v\r\n", err)
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
		fmt.Printf("Ошибка пинга к БД: %v", err)
		return false
	}

	return true
}
