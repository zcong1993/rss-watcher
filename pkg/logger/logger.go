package logger

import (
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/pkg/errors"

	"github.com/sirupsen/logrus"
)

const (
	// DebugLevel has verbose message.
	DebugLevel = "debug"
	// InfoLevel is default log level.
	InfoLevel = "info"
	// WarnLevel is for logging messages about possible issues.
	WarnLevel = "warn"
	// ErrorLevel is for logging errors.
	ErrorLevel = "error"
	// FatalLevel is for logging fatal messages. The system shuts down after logging the message.
	FatalLevel = "fatal"
)

const (
	logFieldTimeStamp = "time"
	logFieldLevel     = "level"
	logFieldScope     = "scope"
	logFieldMessage   = "msg"
)

type Option struct {
	JSONFormatEnabled bool
	ReportCaller      bool
	OutputLevel       string
}

func (o *Option) ApplyDefault() {
	if o.OutputLevel == "" {
		o.OutputLevel = InfoLevel
	}
}

type Logger = logrus.Entry

func NewLogger(name string, option *Option) (*Logger, error) {
	option.ApplyDefault()

	newLogger := logrus.New()
	newLogger.SetOutput(os.Stdout)
	newLogger.SetReportCaller(option.ReportCaller)

	level, err := logrus.ParseLevel(option.OutputLevel)

	if err != nil {
		return nil, errors.Wrap(err, "parse level")
	}

	newLogger.SetLevel(level)

	logger := newLogger.WithField(logFieldScope, name)

	var formatter logrus.Formatter

	fieldMap := logrus.FieldMap{
		logrus.FieldKeyTime:  logFieldTimeStamp,
		logrus.FieldKeyLevel: logFieldLevel,
		logrus.FieldKeyMsg:   logFieldMessage,
	}

	logger.Data = logrus.Fields{
		logFieldScope: logger.Data[logFieldScope],
	}

	if option.JSONFormatEnabled {
		formatter = &logrus.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
			FieldMap:        fieldMap,
		}
	} else {
		formatter = &logrus.TextFormatter{
			TimestampFormat: time.RFC3339Nano,
			FieldMap:        fieldMap,
		}
	}

	logger.Logger.SetFormatter(formatter)

	return logger, nil
}

func BindCobraFlags(app *cobra.Command) func(name string) (*Logger, error) {
	var o Option
	app.PersistentFlags().StringVar(&o.OutputLevel, "log-level", "info", "Options are debug, info, warn, error, or fatal")
	app.PersistentFlags().BoolVar(&o.JSONFormatEnabled, "log-as-json", false, "Print log as JSON (default false)")
	app.PersistentFlags().BoolVar(&o.ReportCaller, "report-caller", false, "Printer caller")

	return func(name string) (*Logger, error) {
		return NewLogger(name, &o)
	}
}
