package log

import (
	"github.com/hhkbp2/go-logging"
	"github.com/sadlil/gologger"
)

var customLogger logging.Logger

var (
	IsDebugEnabled bool
	IsInfoEnabled  bool
	IsWarnEnabled  bool
	IsErrorEnabled bool
	IsFatalEnabled bool
)

func NewLogger(logType string) logging.Logger {
	config_file := "./config/log_config.json"
	if err := logging.ApplyConfigFile(config_file); err != nil {
		Println("Error reading log config file: ", err)
	}
	logger := logging.GetLogger(logType)
	// defer logging.Shutdown()
	return logger
}

func NewTxLogger(file string) gologger.GoLogger {
	return gologger.GetLogger(gologger.FILE, file)
}

func InitLogger(logType string) {
	customLogger = NewLogger(logType)
	InitLogType(logType)
}

func InitLogType(logType string) {
	switch logType {
	case "debug":
		IsDebugEnabled = true
		IsInfoEnabled = true
		IsWarnEnabled = true
		IsErrorEnabled = true
		IsFatalEnabled = true
	case "Info":
		IsInfoEnabled = true
		IsWarnEnabled = true
		IsErrorEnabled = true
		IsFatalEnabled = true
	case "warn":
		IsWarnEnabled = true
		IsErrorEnabled = true
		IsFatalEnabled = true
	case "error":
		IsErrorEnabled = true
		IsFatalEnabled = true
	case "fatal":
		IsFatalEnabled = true
	}
}

func GetLogger() logging.Logger {
	return customLogger
}
