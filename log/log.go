package log

import (
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
)

var Logger logr.Logger

func Init() {
	zapLog, _ := zap.NewDevelopment()

	Logger = zapr.NewLogger(zapLog)
}
