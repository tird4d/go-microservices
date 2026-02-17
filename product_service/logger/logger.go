package logger

import (
	"go.uber.org/zap"
)

var Log *zap.SugaredLogger

func InitLogger(debug bool) error {
	var logger *zap.Logger
	var err error

	if debug {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}

	if err != nil {
		return err
	}

	Log = logger.Sugar()
	return nil
}
