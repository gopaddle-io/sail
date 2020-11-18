package log

import (
	"os"
	"strings"

	"github.com/onrik/logrus/filename"
	"github.com/sirupsen/logrus"
)

var logs *logrus.Logger

func Init() {
	logs = logrus.New()

	//logs.SetFormatter(&logrus.JSONFormatter{})
	logs.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
	})

	//FluentdFormatter
	//logs.SetFormatter(&joonix.FluentdFormatter{})

	logs.SetOutput(os.Stdout)

	logs.SetLevel(logrus.DebugLevel)

	filenameHook := filename.NewHook()
	filenameHook.Field = "source"
	logs.AddHook(filenameHook)
}

func Log(param ...string) *logrus.Entry {
	field := logrus.Fields{}
	field["accountID"] = param[0]
	if len(param) > 1 {
		for _, val := range param[1:] {
			result := strings.Split(val, ":")
			field[result[0]] = result[1]
		}
	}
	return logs.WithFields(field)
}
