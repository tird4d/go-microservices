package logger

import "go.uber.org/zap"

var Log *zap.SugaredLogger

func InitLogger(debug bool) {
	var logger *zap.Logger
	if debug {
		logger, _ = zap.NewDevelopment()
	} else {
		logger, _ = zap.NewProduction()
	}
	Log = logger.Sugar()
}
