package log

import (
	"fmt"
	"log"
	"os"
)

var logger = log.New(os.Stdout, "", log.Lshortfile|log.Ldate|log.Ltime)

//Level 3 - meant for right file name
/// Print
var level = 2

func Print(v ...interface{}) {
	logger.Output(level, fmt.Sprint(v...))
}

func Println(v ...interface{}) {
	logger.Output(level, fmt.Sprintln(v...))
}

func Printf(format string, v ...interface{}) {
	logger.Output(level, fmt.Sprintf(format, v...))
}

func Errorln(v ...interface{}) {
	v = append([]interface{}{"ERROR: "}, v...)
	logger.Output(level, fmt.Sprintln(v...))
}

func Errorf(format string, v ...interface{}) {
	format = "ERROR: " + format
	logger.Output(level, fmt.Sprintf(format, v...))
}

func Warnln(v ...interface{}) {
	v = append([]interface{}{"WARN: "}, v...)
	logger.Output(level, fmt.Sprintln(v...))
}

func Warnf(format string, v ...interface{}) {
	format = "WARN: " + format
	logger.Output(level, fmt.Sprintf(format, v...))
}

/// Panic
func Panic(v ...interface{}) {
	logger.Output(level, fmt.Sprint(v...))
	Panic(1)
}

func Panicln(v ...interface{}) {
	logger.Output(level, fmt.Sprintln(v...))
	panic(1)
}

func Panicf(format string, v ...interface{}) {
	logger.Output(level, fmt.Sprintf(format, v...))
	panic(1)
}

/// Fatal
func Fatal(v ...interface{}) {
	logger.Output(level, fmt.Sprint(v...))
	os.Exit(1)
}

func Fatalln(v ...interface{}) {
	logger.Output(level, fmt.Sprintln(v...))
	os.Exit(1)
}

func Fatalf(format string, v ...interface{}) {
	logger.Output(level, fmt.Sprintf(format, v...))
	os.Exit(1)
}
