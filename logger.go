package gomodule

import (
	"context"
	"io"
	"os"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
)

type Fields logrus.Fields

var loggerInstance *loggerModule

type loggerSettings struct {
	Formatter     string `mapstructure:"formatter"`
	Format        string `mapstructure:"format"`
	File          string `mapstructure:"file"`
	Console       bool   `mapstructure:"console"`
	Color         bool   `mapstructure:"color"`
	Level         string `mapstructure:"level"`
	ReportCaller  bool   `mapstructure:"reportCaller"`
	FilePattern   string `mapstructure:"filePattern"`
	MaxAge        int    `mapstructure:"maxAge"`
	RotationTime  int    `mapstructure:"rotationTime"`
	RotationCount int    `mapstructure:"rotationCount"`
	RotationSize  int    `mapstructure:"rotationSize"`
}

type loggerModule struct {
	DefaultModule
	presettings loggerSettings
	settings    *loggerSettings
	logger      *logrus.Entry
}

func init() {
	loggerInstance = &loggerModule{
		logger: logrus.WithField("module", "logger"),
	}
}

func LoggerModule() IModule {
	return loggerInstance
}

func (l *loggerModule) Logger() *logrus.Entry {
	return l.logger
}

func (l *loggerModule) InitModule(ctx context.Context, _ *Manager) (interface{}, error) {
	l.Logger().Debug("init logger module")
	return &l.presettings, nil
}

func (l *loggerModule) ConfigChanged() {
	if l.settings == nil {
		l.settings = &loggerSettings{}
		*l.settings = l.presettings
		l.reloadSettings()
	} else if *l.settings != l.presettings {
		*l.settings = l.presettings
		l.reloadSettings()
	}
}

func (l *loggerModule) reloadSettings() error {
	if strings.EqualFold(l.settings.Formatter, "text") {
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			ForceColors:     l.settings.Color && l.settings.Console,
			DisableColors:   !l.settings.Color || !l.settings.Console,
			TimestampFormat: l.settings.Format,
		})
	} else {
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: l.settings.Format,
		})
	}

	logrus.SetReportCaller(l.settings.ReportCaller)
	level, err := logrus.ParseLevel(l.settings.Level)
	if err != nil {
		return err
	}
	logrus.SetLevel(level)
	var writer *rotatelogs.RotateLogs

	filePattern := l.settings.File
	if l.settings.FilePattern != "" {
		filePattern += "." + l.settings.FilePattern
	}

	logrus.Info("filePattern ", filePattern)
	if l.settings.MaxAge > 0 {
		writer, err = rotatelogs.New(
			filePattern,
			rotatelogs.WithLinkName(l.settings.File),
			rotatelogs.WithMaxAge(time.Duration(l.settings.MaxAge)*time.Hour),
			rotatelogs.WithRotationSize(int64(l.settings.RotationSize)*1024*1024),
			rotatelogs.WithRotationTime(time.Duration(l.settings.RotationTime)*time.Hour),
		)
		if err != nil {
			return err
		}
	} else if l.settings.RotationCount > 0 {
		writer, err = rotatelogs.New(
			filePattern,
			rotatelogs.WithLinkName(l.settings.File),
			rotatelogs.WithRotationCount(uint(l.settings.RotationCount)),
			rotatelogs.WithRotationSize(int64(l.settings.RotationSize)*1024*1024),
			rotatelogs.WithRotationTime(time.Duration(l.settings.RotationTime)*time.Hour),
		)
		if err != nil {
			return err
		}
	}

	var output io.Writer
	if l.settings.Console {
		output = io.MultiWriter(writer, os.Stdout)
	} else {
		output = io.MultiWriter(writer)
	}

	logrus.SetOutput(output)

	return nil
}
