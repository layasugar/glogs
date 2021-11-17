// Package glogs is a global internal glogs
// glogs: this is extend package, use https://github.com/uber-go/zap
package glogs

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"net/http"
	"os"
)

// InitLog 初始化日志文件 logPath= /home/logs/app/appName/childPath
func InitLog(options ...LogOptionFunc) {
	for _, f := range options {
		f(DefaultConfig)
	}

	Sugar = initSugar(DefaultConfig)
}

func initSugar(lc *Config) *zap.Logger {
	loglevel := zapcore.InfoLevel
	defaultLogLevel.SetLevel(loglevel)

	logPath := fmt.Sprintf("%s/%s/%s", lc.logPath, lc.appName, lc.childPath)

	var core zapcore.Core
	// 打印至文件中
	if lc.logType == "file" {
		configs := zap.NewProductionEncoderConfig()
		configs.FunctionKey = "func"
		configs.EncodeTime = timeEncoder

		w := zapcore.AddSync(GetWriter(logPath, lc))

		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(configs),
			w,
			defaultLogLevel,
		)
		log.Printf("[glogs_sugar] log success")
	} else {
		// 打印在控制台
		consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		core = zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), defaultLogLevel)
		log.Printf("[glogs_sugar] log success")
	}

	filed := zap.Fields(zap.String("app_name", lc.appName), zap.String("app_mode", lc.appMode))
	return zap.New(core, filed, zap.AddCaller(), zap.AddCallerSkip(3))
	//Sugar = logger.Sugar()
}

func Info(template string, args ...interface{}) {
	msg, fields := dealWithArgs(template, args...)
	writer(nil, Sugar, LevelInfo, msg, LevelInfo, fields...)
}
func InfoF(r *http.Request, title string, template string, args ...interface{}) {
	msg, fields := dealWithArgs(template, args...)
	writer(r, Sugar, LevelInfo, msg, title, fields...)
}

func Warn(template string, args ...interface{}) {
	msg, fields := dealWithArgs(template, args...)
	writer(nil, Sugar, LevelWarn, msg, LevelWarn, fields...)
}
func WarnF(r *http.Request, title string, template string, args ...interface{}) {
	msg, fields := dealWithArgs(template, args...)
	writer(r, Sugar, LevelWarn, msg, title, fields...)
}

func Error(template string, args ...interface{}) {
	msg, fields := dealWithArgs(template, args...)
	writer(nil, Sugar, LevelError, msg, LevelError, fields...)
}
func ErrorF(r *http.Request, title string, template string, args ...interface{}) {
	msg, fields := dealWithArgs(template, args...)
	writer(r, Sugar, LevelError, msg, title, fields...)
}
