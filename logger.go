package gomodule

import (
	"context"
	"io"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

type Fields logrus.Fields

var loggerInstance *loggerModule

type loggerSettings struct {
	Formatter    string `mapstructure:"formatter"`
	Format       string `mapstructure:"format"`
	File         string `mapstructure:"file"`
	Console      bool   `mapstructure:"console"`
	Level        string `mapstructure:"level"`
	ReportCaller bool   `mapstructure:"reportCaller"`
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

func (l *loggerModule) InitModule(ctx context.Context, wg *sync.WaitGroup) (interface{}, error) {
	l.Logger().Infof("init logger module")
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
			ForceColors:     true,
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

	fd, err := createLogFile(l.settings.File)
	if err != nil {
		return err
	}

	var output io.Writer
	if l.settings.Console {
		output = io.MultiWriter(fd, os.Stdout)
	} else {
		output = io.MultiWriter(fd)
	}

	logrus.SetOutput(output)

	return nil
}

func createLogFile(file string) (*os.File, error) {
	dir := path.Dir(file)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return nil, err
		}
	}

	fd, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return fd, nil
}
