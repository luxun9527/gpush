package pkg

import (
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"

	"time"
)

func getLevel(level string) zapcore.Level {
	switch level {
	case zapcore.DebugLevel.String():
		return zapcore.DebugLevel
	case zapcore.InfoLevel.String():
		return zapcore.InfoLevel
	case zapcore.WarnLevel.String():
		return zapcore.WarnLevel
	case zapcore.ErrorLevel.String():
		return zapcore.ErrorLevel
	case zapcore.InfoLevel.String():
		return zapcore.DPanicLevel
	case zapcore.InfoLevel.String():
		return zapcore.PanicLevel
	case zapcore.InfoLevel.String():
		return zapcore.FatalLevel
	default:
		return zapcore.Level(100)
	}

}

type LoggerConfig struct {
	//日志级别
	Level string
	//当错误时，是否显示堆栈
	Stacktrace bool
	//添加调用者信息
	AddCaller bool
	//是否控制台显示，/dev/stdout 与输入文件互斥
	Debug bool

	//文件名称加路径
	FileName string
	//warn 级别的日志输出到不同的地方
	WarnFileName string
	// 日志轮转大小，单位MB，默认500MB
	MaxSize int32
	//日志轮转最大时间，单位day，默认1 day
	MaxAge int32
	//日志轮转个数，默认10
	MaxBackup int32
	//日志轮转周期，默认24 hour
	Interval int32
	//异步日志
	Async bool
	//是否 输出json格式的数据，JSON格式相对于console格式，不方便阅读，但是对机器更加友好
	Json bool
}

func (c *LoggerConfig) Build() *zap.Logger {
	//
	var (
		ws      zapcore.WriteSyncer
		warnWs  zapcore.WriteSyncer
		encoder zapcore.Encoder
	)

	// First, define our level-handling logic.
	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.WarnLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.WarnLevel

	})
	encoderConfig := zapcore.EncoderConfig{
		//当存储的格式为JSON的时候这些作为可以key
		MessageKey:    "message",
		LevelKey:      "level",
		TimeKey:       "time",
		NameKey:       "logger",
		CallerKey:     "caller",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		//以上字段输出的格式
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     CustomTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
	}

	if c.Debug {
		ws = zapcore.Lock(os.Stdout)
		warnWs = zapcore.Lock(os.Stderr)
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		nomalConfig := &lumberjack.Logger{
			Filename:   c.FileName,
			MaxSize:    int(c.MaxSize),
			MaxAge:     int(c.MaxAge),
			MaxBackups: int(c.MaxBackup),
			LocalTime:  true,
			Compress:   true,
		}
		warnConfig := &lumberjack.Logger{
			Filename:   c.WarnFileName,
			MaxSize:    int(c.MaxSize),
			MaxAge:     int(c.MaxAge),
			MaxBackups: int(c.MaxBackup),
			LocalTime:  true,
			Compress:   true,
		}
		ws = zapcore.AddSync(nomalConfig)
		warnWs = zapcore.AddSync(warnConfig)
	}
	if c.Json {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}
	if getLevel(c.Level) == 100 {
		log.Panicf("error level %v", c.Level)
	}
	highCore := zapcore.NewCore(encoder, warnWs, highPriority)
	lowCore := zapcore.NewCore(encoder, ws, lowPriority)
	core := zapcore.NewTee(highCore, lowCore)
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

}

func CustomTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 - 15:04:05"))
}
