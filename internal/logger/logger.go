package logger

import (
	"go.uber.org/zap"
)

var Log *zap.Logger = zap.NewNop()
var LogSugar *zap.SugaredLogger

func init() {
	logger, _ := zap.NewProduction()
	Log = logger
	defer logger.Sync()

	LogSugar = logger.Sugar()
}
