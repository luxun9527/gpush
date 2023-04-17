package config

import (
	"go.uber.org/zap"
	"log"
	"testing"
)

func TestLogger(t *testing.T) {
	log.Println(zap.DebugLevel)
	l:=LoggerConfig{
		Level:      "debug",
		Stacktrace: true,
		AddCaller:  true,
		Debug:      false,
		FileName:   "./stdout.ws.json",
		WarnFileName:   "./stderr.ws.json",
		MaxSize:    100,
		MaxAge:     10,
		MaxBackup:  10,
		Interval:   0,
		Async:      false,
		Json:       true,
	}
	logger:=l.Build()
	logger.Info("test",zap.String("test","test"))
	logger.Debug("test")
	logger.Warn("test")
	logger.Error("test")
	logger.Panic("test")
}
