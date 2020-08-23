package logger

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"time"
)

var DefaultLogger *zap.Logger
var SugarLogger *zap.SugaredLogger

// env: online/dev, defaults to dev
func InitLogger(env string) (err error) {
	switch env {
	case "online":
		rawJson := []byte(`{
"level": "info",
"encoding": "console",
"outputPaths": ["stdout", "./logs/logs"],
"errorOutputPaths": ["stderr"],
"initialFields": {},
"encoderConfig": {
	"messageKey": "m",
	"levelKey": "lv",
	"levelEncoder": "capital",
	"callerKey": "c",
	"callerEncoder": "",
	"timeKey": "t",
	"timeEncoder": "iso8601",
	"nameKey": "n",
	"nameEncoder": "",
	"durationKey": "d",
	"durationEncoder": "string"
}
		}`)
		var cfg zap.Config
		if err = json.Unmarshal(rawJson, &cfg); err != nil {
			panic(err)
		}
		DefaultLogger, err = cfg.Build(zap.AddCallerSkip(1))
		if err != nil {
			fmt.Println(err)
			return
		}
	case "dev":
		fallthrough
	default:
		env = "dev"
		rawJson := []byte(`{
"level": "debug",
"encoding": "console",
"outputPaths": ["stdout", "./logs/logs"],
"errorOutputPaths": ["stderr"],
"initialFields": {},
"encoderConfig": {
	"messageKey": "m",
	"levelKey": "lv",
	"levelEncoder": "capital",
	"callerKey": "c",
	"callerEncoder": "",
	"timeKey": "t",
	"timeEncoder": "iso8601",
	"nameKey": "n",
	"nameEncoder": "",
	"durationKey": "d",
	"durationEncoder": "string"
}
		}`)
		var cfg zap.Config
		if err = json.Unmarshal(rawJson, &cfg); err != nil {
			panic(err)
		}
		DefaultLogger, err = cfg.Build(zap.AddCallerSkip(1))
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	SugarLogger = DefaultLogger.Sugar()

	return
}

// before exits program, call it to print buffered msgs.
func Exit() {
	DefaultLogger.Sync()
	SugarLogger.Sync()
	return
}

type Field = zap.Field

func Duration(key string, dur time.Duration) Field {
	return Field(zap.Duration(key, dur))
}
func Any(key string, f interface{}) Field {
	return Field(zap.Any(key, f))
}

// following logging methods have great efficiency
func Debug(msg string, fields ...Field) {
	if DefaultLogger == nil {
		return
	}
	DefaultLogger.Debug(msg, fields...)
}
func Info(msg string, fields ...Field) {
	if DefaultLogger == nil {
		return
	}
	DefaultLogger.Info(msg, fields...)
}
func Warn(msg string, fields ...Field) {
	if DefaultLogger == nil {
		return
	}
	DefaultLogger.Warn(msg, fields...)
}
func Error(msg string, fields ...Field) {
	if DefaultLogger == nil {
		return
	}
	DefaultLogger.Error(msg, fields...)
}
func Panic(msg string, fields ...Field) {
	if DefaultLogger == nil {
		return
	}
	DefaultLogger.Panic(msg, fields...)
}
func Fatal(msg string, fields ...Field) {
	if DefaultLogger == nil {
		return
	}
	DefaultLogger.Fatal(msg, fields...)
}

// f/sugar loggers uses fmt.Sprintf
// degraded efficiency, use at non-hot-path
func Debugf(template string, args ...interface{}) {
	if SugarLogger == nil {
		return
	}
	SugarLogger.Debugf(template, args...)
	return
}
func Infof(template string, args ...interface{}) {
	if SugarLogger == nil {
		return
	}
	SugarLogger.Infof(template, args...)
	return
}
func Warnf(template string, args ...interface{}) {
	if SugarLogger == nil {
		return
	}
	SugarLogger.Warnf(template, args...)
	return
}
func Errorf(template string, args ...interface{}) {
	if SugarLogger == nil {
		return
	}
	SugarLogger.Errorf(template, args...)
	return
}
func Panicf(template string, args ...interface{}) {
	if SugarLogger == nil {
		return
	}
	SugarLogger.Panicf(template, args...)
	return
}
func Fatalf(template string, args ...interface{}) {
	if SugarLogger == nil {
		return
	}
	SugarLogger.Fatalf(template, args...)
	return
}
