package log

import (
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

var (
	loggers = []xLogger{
		newDefault(os.Stdout),
	}

	style = newStyles()
)

const TAG = "OMEGA::"

type xLogger struct {
	*log.Logger
	source *os.File
}

func newStyles() *log.Styles {
	styles := log.DefaultStyles()

	style := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1, 0, 1)

	styles.Levels[log.DebugLevel] = style.
		SetString(TAG + "D").
		Background(lipgloss.Color("#414868")).
		Foreground(lipgloss.Color("#FFFFFF"))

	styles.Levels[log.InfoLevel] = style.
		SetString(TAG + "I").
		Background(lipgloss.Color("#485E30")).
		Foreground(lipgloss.Color("#FFFFFF"))

	styles.Levels[log.WarnLevel] = style.
		SetString(TAG + "W").
		Background(lipgloss.Color("#FF9E64")).
		Foreground(lipgloss.Color("#000000"))

	styles.Levels[log.ErrorLevel] = style.
		SetString(TAG + "E").
		Background(lipgloss.Color("#F7768E")).
		Foreground(lipgloss.Color("#000000"))

	styles.Levels[log.FatalLevel] = style.
		SetString(TAG + "F").
		Background(lipgloss.Color("#BB9AF7")).
		Foreground(lipgloss.Color("#000000"))

	return styles

}

func new(w io.Writer) *log.Logger {
	logger := log.NewWithOptions(w, log.Options{
		ReportTimestamp: true,
		Level:           log.DebugLevel,
	})
	logger.SetStyles(style)
	return logger
}

func newDefault(w io.Writer) xLogger {
	return xLogger{
		Logger: new(w),
		source: nil,
	}
}

func InitFileLogger(filePath string) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	loggers = append(loggers, xLogger{
		Logger: new(file),
		source: file,
	})
	return nil
}

func CloseAll() []error {
	errors := make([]error, 0, len(loggers))
	for _, l := range loggers {
		if l.source != nil {
			err := l.source.Close()
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

	switch strings.ToLower(str) {
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
