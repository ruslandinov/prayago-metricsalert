package dbstorage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type (
	DBStorageConfig struct {
		ConnectionString string
	}

	DBStorage struct {
		config DBStorageConfig
		db     *sql.DB
	}

	DBStorager interface {
		Close()
		Ping() bool
	}
)

func NewDBStorage(config DBStorageConfig) DBStorager {
	fmt.Printf("NewDBStorage config: %v\r\n", config)
	db, err := sql.Open("pgx", config.ConnectionString)
	if err != nil {
		fmt.Printf("Ошибка подключения к БД: %v\r\n", err)
	}

	dbstorage := &DBStorage{
		config,
		db,
	}

	return dbstorage
}

func (dbs *DBStorage) Close() {
	dbs.db.Close()
}

func (dbs *DBStorage) Ping() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := dbs.db.PingContext(ctx); err != nil {
		fmt.Printf("Ошибка пинга к БД: %v", err)
		return false
	}

	return true
}
