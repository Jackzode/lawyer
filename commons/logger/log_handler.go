package glog

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var Klog *zap.Logger
var Slog *zap.SugaredLogger

func InitLogger(filename string, level string) {

	encoder := getEncoder()
	writeSyncer := getLogWriter(filename)
	zlevel := zapcore.DebugLevel
	switch level {
	case "info":
		zlevel = zapcore.InfoLevel
	case "error":
		zlevel = zapcore.ErrorLevel
	}
	core := zapcore.NewCore(encoder, writeSyncer, zlevel)

	Klog = zap.New(core, zap.AddCaller())
	Slog = Klog.Sugar()
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // 修改时间编码器
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter(filename string) zapcore.WriteSyncer {
	file, _ := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, os.ModeAppend)
	return zapcore.AddSync(file)
}
