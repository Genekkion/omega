package log

import (
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/charmbracelet/log"
)

var (
	loggers = []Logger{
		new(os.Stdout),
	}
)

const TAG = "OMEGA::"

type Logger struct {
	*log.Logger
	deconstructor func() error
}

func new(w io.Writer) Logger {
	l := log.NewWithOptions(w, log.Options{
		ReportTimestamp: true,
		Level:           log.DebugLevel,
	})
	l.SetStyles(defaultStyle)
	return Logger{Logger: l}
}

func newWithDec(w io.Writer, deconstructor func() error) Logger {
	l := new(w)
	l.deconstructor = deconstructor
	return l
}

// Assumes file checks occurred at a higher level
func NewFromFile(path string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	loggers = append(loggers, newWithDec(file, file.Close))
	return nil
}

func CloseAll() []error {
	errors := make([]error, 0, len(loggers))
	for _, l := range loggers {
		if l.deconstructor != nil {
			err := l.deconstructor()
			if err != nil {
				errors = append(errors, err)
			}
		}
	}
	return errors
}

func SetLogLevel(str string) {
	const defaultLevel = log.InfoLevel
	level, err := log.ParseLevel(str)
	if err != nil {
		Warn("Invalid log level specified, using default log level", "level", defaultLevel.String())
		level = defaultLevel
	}
	for _, l := range loggers {
		l.SetLevel(level)
	}

	switch strings.ToLower(strings.TrimSpace(str)) {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
}

// Wrapper function to log errors if they exist
func ErrorWrapper(err error) error {
	if err != nil {
		pc, file, line, ok := runtime.Caller(1)
		for _, l := range loggers {
			if !ok {
				l.Error("Error retrieving stack trace.")
			} else {
				l.Error("Stack trace:",
					"function", runtime.FuncForPC(pc).Name(),
					"file", file,
					"line", line,
				)
			}
			l.Error("Program received error", "err", err)
		}
	}
	return err
}

func FatalWrapper(err error) error {
	if err != nil {
		pc, file, line, ok := runtime.Caller(1)
		for _, l := range loggers {
			if !ok {
				l.Error("Error retrieving stack trace.")
			} else {
				l.Error("Stack trace:",
					"function", runtime.FuncForPC(pc).Name(),
					"file", file,
					"line", line,
				)
			}
			l.Fatal("Program received fatal error", "err", err)
		}
	}
	return err
}

func DebugCaller(message interface{}, keyValue ...interface{}) {
	for _, l := range loggers {
		l.Debug(message, keyValue...)
		pc, file, line, ok := runtime.Caller(1)
		if !ok {
			l.Debug("Error retrieving stack trace.")
		} else {
			l.Debug("Stack trace:",
				"function", runtime.FuncForPC(pc).Name(),
				"file", file,
				"line", line,
			)
		}
	}
}

func Error(message interface{}, keyValue ...interface{}) {
	for _, l := range loggers {
		l.Error(message, keyValue...)
	}
}

func Fatal(message interface{}, keyValue ...interface{}) {
	for _, l := range loggers {
		l.Fatal(message, keyValue...)
	}
}

func Warn(message interface{}, keyValue ...interface{}) {
	for _, l := range loggers {
		l.Warn(message, keyValue...)
	}
}

func Debug(message interface{}, keyValue ...interface{}) {
	for _, l := range loggers {
		l.Debug(message, keyValue...)
	}
}

func Info(message interface{}, keyValue ...interface{}) {
	for _, l := range loggers {
		l.Info(message, keyValue...)
	}
}
