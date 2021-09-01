package glogs

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"gorm.io/gorm/logger"
	"time"

	"gorm.io/gorm/utils"
)

var SqlLogger = "sql_logger"

func Default(writer *zap.Logger, level logger.LogLevel) logger.Interface {
	var config = logger.Config{
		SlowThreshold: 200 * time.Millisecond,
		LogLevel:      level,
		Colorful:      true,
	}
	var (
		infoStr      = "{\"line\": \"%s\", \"level\": \"[info]\", \"msg\": \"%s\"}"
		warnStr      = "{\"line\": \"%s\", \"level\": \"[warn]\", \"msg\": \"%s\"}"
		errStr       = "{\"line\": \"%s\", \"level\": \"[error]\", \"msg\": \"%s\"}"
		traceStr     = "{\"line\": \"%s\", \"耗时\": \"%.3fms\", \"rows\": \"%v\" ,\"sql\": \"%s\"}"
		traceWarnStr = "{\"line\": \"%s\", \"错误\": \"%s\", \"耗时\": \"%.3fms\", \"rows\": \"%v\", \"sql\": \"%s\"}"
		traceErrStr  = "{\"line\": \"%s\", \"slow\": \"%s\", \"耗时\": \"%.3fms\", \"rows\": \"%v\", \"sql\": \"%s\"}"
	)

	return &gormLogger{
		Writer:       writer,
		Config:       config,
		infoStr:      infoStr,
		warnStr:      warnStr,
		errStr:       errStr,
		traceStr:     traceStr,
		traceWarnStr: traceWarnStr,
		traceErrStr:  traceErrStr,
	}
}

type gormLogger struct {
	Writer *zap.Logger
	logger.Config
	infoStr, warnStr, errStr            string
	traceStr, traceErrStr, traceWarnStr string
}

// LogMode log mode
func (l *gormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

// Info print info
func (l *gormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		errInfo := fmt.Sprintf(msg, data)
		writer(ctx, l.Writer, LevelInfo, fmt.Sprintf(l.infoStr, utils.FileWithLineNum(), errInfo), SqlLogger)
	}
}

// Warn print warn messages
func (l *gormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		errInfo := fmt.Sprintf(msg, data)
		writer(ctx, l.Writer, LevelWarn, fmt.Sprintf(l.infoStr, utils.FileWithLineNum(), errInfo), SqlLogger)
	}
}

// Error print error messages
func (l *gormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		errInfo := fmt.Sprintf(msg, data)
		writer(ctx, l.Writer, LevelError, fmt.Sprintf(l.infoStr, utils.FileWithLineNum(), errInfo), SqlLogger)
	}
}

// Trace print sql message
func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel > 0 {
		elapsed := time.Since(begin)
		switch {
		case err != nil && l.LogLevel >= logger.Error:
			sql, rows := fc()
			if rows == -1 {
				msg := fmt.Sprintf(l.traceErrStr, utils.FileWithLineNum(), err.Error(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
				writer(ctx, l.Writer, LevelError, msg, SqlLogger)
			} else {
				msg := fmt.Sprintf(l.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
				writer(ctx, l.Writer, LevelError, msg, SqlLogger)
			}
		case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn:
			sql, rows := fc()
			slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
			if rows == -1 {
				msg := fmt.Sprintf(l.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
				writer(ctx, l.Writer, LevelWarn, msg, SqlLogger)
			} else {
				msg := fmt.Sprintf(l.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
				writer(ctx, l.Writer, LevelWarn, msg, SqlLogger)
			}
		case l.LogLevel >= logger.Info:
			sql, rows := fc()
			if rows == -1 {
				msg := fmt.Sprintf(l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
				writer(ctx, l.Writer, LevelInfo, msg, SqlLogger)
			} else {
				msg := fmt.Sprintf(l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
				writer(ctx, l.Writer, LevelInfo, msg, SqlLogger)
			}
		}
	}
}
