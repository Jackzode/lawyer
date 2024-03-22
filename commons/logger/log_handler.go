package glog

import (
	"go.uber.org/zap"
)

var Logger *zap.Logger

func InitLog(logFile string) {

	config := zap.NewProductionConfig()
	config.OutputPaths = []string{logFile} //"logfile.log"
	Logger, _ = config.Build()
}
