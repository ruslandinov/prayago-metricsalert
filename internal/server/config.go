package server

import (
	"flag"
	"os"
	"prayago-metricsalert/internal/logger"
	"strconv"
	"time"
)

type ServerConfig struct {
	ServerAddress         string
	StorageFPath          string
	DBConnectionString    string
	StoreInterval         time.Duration
	RestoreStorageOnStart bool
}

func NewServerConfig() ServerConfig {
	a := flag.String("a", "localhost:8080", "server address and port")
	f := flag.String("f", "./storage.json", "memstorage file path")
	d := flag.String("d", "", "database connection string")
	i := flag.Int("i", 300, "memstorage saving interval, sec")
	r := flag.Bool("r", true, "should memstorage restore values on server start or not")
	flag.Parse()

	config := ServerConfig{
		ServerAddress: *a,
		StorageFPath: *f,
		DBConnectionString: *d,
		StoreInterval: time.Duration(*i) * time.Second,
		RestoreStorageOnStart: *r,
	}

	if envServerAddress := os.Getenv("ADDRESS"); envServerAddress != "" {
		config.ServerAddress = envServerAddress
	}
	if envStorageFPath := os.Getenv("FILE_STORAGE_PATH"); envStorageFPath != "" {
		config.StorageFPath = envStorageFPath
	}
	if envStoreIntrvl := os.Getenv("STORE_INTERVAL"); envStoreIntrvl != "" {
		if storeIntervalInt, err := strconv.Atoi(envStoreIntrvl); err == nil {
			config.StoreInterval = time.Duration(storeIntervalInt) * time.Second
		}
	}
	if envRestoreStorageOnStart := os.Getenv("RESTORE"); envRestoreStorageOnStart != "" {
		config.RestoreStorageOnStart, _ = strconv.ParseBool(envRestoreStorageOnStart)
	}
	if envDataBaseDsn := os.Getenv("DATABASE_DSN"); envDataBaseDsn != "" {
		config.DBConnectionString = envDataBaseDsn
	}

	logger.LogSugar.Infoln("Server config:", config)

	return config
}
