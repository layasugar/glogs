package glogs

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type CusLog struct {
	Logger *zap.Logger
	Config *LogConfig
}

// NewLogger 得到一个zap.Logger
func NewLogger(options ...LogOptionFunc) *CusLog {
	var cus = &CusLog{Config: defaultLogConfig}
	for _, f := range options {
		f(cus.Config)
	}

	cus.Logger = initSugar(cus.Config)
	return cus
}

func (l *CusLog) Info(template string, args ...interface{}) {
	msg, fields := dealWithArgs(template, args...)
	writer(context.Background(), l.Logger, LevelInfo, msg, LevelInfo, fields...)
}
func (l *CusLog) InfoF(c *gin.Context, title string, template string, args ...interface{}) {
	msg, fields := dealWithArgs(template, args...)
	writer(c, l.Logger, LevelInfo, msg, title, fields...)
}

func (l *CusLog) Warn(template string, args ...interface{}) {
	msg, fields := dealWithArgs(template, args...)
	writer(context.Background(), l.Logger, LevelWarn, msg, LevelWarn, fields...)
}
func (l *CusLog) WarnF(c *gin.Context, title string, template string, args ...interface{}) {
	msg, fields := dealWithArgs(template, args...)
	writer(c, l.Logger, LevelWarn, msg, title, fields...)
}

func (l *CusLog) Error(template string, args ...interface{}) {
	msg, fields := dealWithArgs(template, args...)
	writer(context.Background(), l.Logger, LevelError, msg, LevelError, fields...)
}
func (l *CusLog) ErrorF(c *gin.Context, title string, template string, args ...interface{}) {
	msg, fields := dealWithArgs(template, args...)
	writer(c, l.Logger, LevelError, msg, title, fields...)
}
