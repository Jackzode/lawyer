package glog

import (
	"go.uber.org/zap"
)

var Logger *zap.Logger

func InitLog(logFile string) {
	// 初始化全局日志处理器
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{logFile} //"logfile.log"
	Logger, _ = config.Build()
}
