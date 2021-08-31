package glogs

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	rl "github.com/layasugar/glogs/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"time"
)

const (
	RequestIDName    = "x-b3-traceid"
	HeaderAppName    = "app-name"
	KeyPath          = "path"
	KeyTitle         = "title"
	KeyOriginAppName = "origin_app_name"
	KeyCtx           = "ctx"

	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
)

type LogConfig struct {
	appName       string        // 应用名
	appMode       string        // 应用环境
	logType       string        // 日志类型
	logPath       string        // 日志主路径
	childPath     string        // 日志子路径+文件名
	RotationSize  int64         // 单个文件大小
	RotationCount uint          // 可以保留的文件个数
	NoBuffWrite   bool          // 设置无缓冲日志写入
	RotationTime  time.Duration // 日志分割的时间
	MaxAge        time.Duration // 日志最大保留的天数
}

type LogOptionFunc func(*LogConfig)

var (
	Sugar *zap.Logger

	defaultLogLevel  = zap.NewAtomicLevel()
	defaultLogConfig = &LogConfig{
		appName:      "default-app",        // 默认应用名称
		appMode:      "dev",                // 默认应用环境
		logType:      "file",               // 默认日志类型
		logPath:      "/home/logs/app",     // 默认文件目录
		childPath:    "glogs/%Y-%m-%d.log", // 默认子目录
		RotationSize: 32 * 1024 * 1024,     // 默认大小为32M
	}
	SqlLogger = "sql_logger"
)

// 设置应用名称,默认值default-app
func SetLogAppName(appName string) LogOptionFunc {
	return func(c *LogConfig) {
		if appName != "" {
			c.appName = appName
		}
	}
}

// 设置环境变量,标识当前应用运行的环境,默认值dev
func SetLogAppMode(appMode string) LogOptionFunc {
	return func(c *LogConfig) {
		if appMode != "" {
			c.appMode = appMode
		}
	}
}

// 设置日志类型,日志类型目前分为2种,console和file,默认值file
func SetLogType(logType string) LogOptionFunc {
	return func(c *LogConfig) {
		if logType != "" {
			c.logType = logType
		}
	}
}

// 设置日志目录,这个是主目录,程序会给此目录拼接上项目名,子目录以及文件,默认值/home/logs/app
func SetLogPath(logPath string) LogOptionFunc {
	return func(c *LogConfig) {
		if logPath != "" {
			c.logPath = logPath
		}
	}
}

// 设置子目录—+文件名,保证一个类型的文件在同一个文件夹下面便于区分、默认值glogs/%Y-%m-%d.log
func SetLogChildPath(childPath string) LogOptionFunc {
	return func(c *LogConfig) {
		if childPath != "" {
			c.childPath = childPath
		}
	}
}

// 设置单个文件最大值byte,默认值32M
func SetLogMaxSize(size int64) LogOptionFunc {
	return func(c *LogConfig) {
		if size > 0 {
			c.RotationSize = size
		}
	}
}

// 设置文件最大保留时间、默认值7天
func SetLogMaxAge(maxAge time.Duration) LogOptionFunc {
	return func(c *LogConfig) {
		if maxAge != 0 {
			c.MaxAge = maxAge
		}
	}
}

// 设置文件分割时间、默认值24*time.Hour(按天分割)
func SetRotationTime(rotationTime time.Duration) LogOptionFunc {
	return func(c *LogConfig) {
		if rotationTime != 0 {
			c.RotationTime = rotationTime
		}
	}
}

// 设置保留的最大文件数量、没有默认值(表示不限制)
func SetRotationCount(n uint) LogOptionFunc {
	return func(c *LogConfig) {
		if n != 0 {
			c.RotationCount = n
		}
	}
}

// 设置无缓冲写入日志，比较消耗性能
func SetNoBuffWriter() LogOptionFunc {
	return func(c *LogConfig) {
		c.NoBuffWrite = true
	}
}

func writer(ctx context.Context, logger *zap.Logger, level, msg string, title string, fields ...zap.Field) {
	if logger == nil {
		fmt.Println(msg)
		return
	}

	// 断言ctx是什么类型
	switch ctx.(type) {
	case *gin.Context:
		if c, ok := ctx.(*gin.Context); ok {
			requestID := c.GetHeader(RequestIDName)
			originAppName := c.GetHeader(HeaderAppName)
			path := c.Request.RequestURI
			fields = append(fields, zap.String(KeyPath, path),
				zap.String(RequestIDName, requestID),
				zap.String(KeyTitle, title),
				zap.String(KeyOriginAppName, originAppName))
			do(logger, level, msg, fields...)
		}
	case nil:
		fields = append(fields, zap.String(KeyTitle, title))
		do(logger, level, msg, fields...)
	default:
		s := fmt.Sprintf("%s", ctx)
		fields = append(fields, zap.String(KeyCtx, s), zap.String(KeyTitle, title))
		do(logger, level, msg, fields...)
	}
}

func do(logger *zap.Logger, level, msg string, fields ...zap.Field) {
	switch level {
	case LevelInfo:
		logger.Info(msg, fields...)
	case LevelWarn:
		logger.Warn(msg, fields...)
	case LevelError:
		logger.Error(msg, fields...)
	}
}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	var layout = "2006-01-02 15:04:05"
	type appendTimeEncoder interface {
		AppendTimeLayout(time.Time, string)
	}

	if enc, ok := enc.(appendTimeEncoder); ok {
		enc.AppendTimeLayout(t, layout)
		return
	}

	enc.AppendString(t.Format(layout))
}

// 按天切割按大小切割
// filename 文件名
// RotationSize 每个文件的大小
// MaxAge 文件最大保留天数
// RotationCount 最大保留文件个数
// RotationTime 设置文件分割时间
// RotationCount 设置保留的最大文件数量
func GetWriter(filename string, lc *LogConfig) io.Writer {
	// 生成rotatelogs的Logger 实际生成的文件名 stream-2021-5-20.log
	// demo.log是指向最新日志的连接
	// 保存7天内的日志，每1小时(整点)分割一第二天志
	var options []rl.Option
	if lc.NoBuffWrite {
		options = append(options, rl.WithNoBuffer())
	}
	options = append(options,
		rl.WithRotationSize(lc.RotationSize),
		rl.WithRotationCount(lc.RotationCount),
		rl.WithRotationTime(lc.RotationTime),
		rl.WithMaxAge(lc.MaxAge),
		rl.ForceNewFile())

	hook, err := rl.New(
		filename,
		options...,
	)

	if err != nil {
		panic(err)
	}
	return hook
}

func dealWithArgs(tmp string, args ...interface{}) (msg string, f []zap.Field) {
	var tmpArgs []interface{}
	for _, item := range args {
		if zapField, ok := item.(zap.Field); ok {
			f = append(f, zapField)
		} else {
			tmpArgs = append(tmpArgs, item)
		}
	}
	msg = fmt.Sprintf(tmp, tmpArgs...)
	return
}

func String(key string, value interface{}) zap.Field {
	v := fmt.Sprintf("%v", value)
	return zap.String(key, v)
}
