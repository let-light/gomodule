package main

import (
	"context"
	"sync"
	"time"

	"github.com/let-light/gomodule"
	"github.com/let-light/gomodule/examples/configcenter"
	feature_configcenter "github.com/let-light/gomodule/examples/features/configcenter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type MainFlags struct {
}

type Settings struct {
	Test string `mapstructure:"test"`
}

type SimpleModule struct {
	gomodule.DefaultModule
	flags       *MainFlags
	presettings Settings
	settings    *Settings
	ctx         context.Context
	logger      *logrus.Entry
	mutex       sync.RWMutex
}

var instance *SimpleModule

func init() {
	instance = &SimpleModule{
		flags:    &MainFlags{},
		settings: &Settings{},
		logger:   logrus.WithField("module", "simple"),
	}
}

func (s *SimpleModule) Logger() *logrus.Entry {
	return instance.logger
}

func (s *SimpleModule) InitCommand() ([]*cobra.Command, error) {
	s.Logger().Info("init command")
	cmd := &cobra.Command{
		Use:   "desc",
		Short: "simple module",
		Run: func(cmd *cobra.Command, args []string) {
			for {
				select {
				case <-s.ctx.Done():
					s.Logger().Info("simple module command done")
					return
				case <-time.After(time.Second):
					s.Logger().Info("simple module command run ...")
				}
			}
		},
	}

	return []*cobra.Command{cmd}, nil
}

func (s *SimpleModule) InitModule(ctx context.Context, m *gomodule.Manager) (interface{}, error) {
	s.ctx = ctx
	s.Manager = m

	m.RequireFeatures(func(cc feature_configcenter.Feature, ss *SimpleModule) {
		cc.HelloWorld()
		ss.Logger().Info("require modules done")
	})
	s.Logger().Info("init simple module")
	return &s.presettings, nil
}

func (s *SimpleModule) ConfigChanged() {
	s.Logger().Info("simple module config changed")
	if err := s.settings == nil; err {
		s.settings = &Settings{}
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	*s.settings = s.presettings

	s.Logger().Info("simple module config changed done")
}

func (s *SimpleModule) SafeSettings() Settings {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return *s.settings
}

func (s *SimpleModule) ModuleRun() {
	s.Logger().Info("simple module run ...")

	s.Logger().Info("settings: ", s.settings.Test)
	for {
		select {
		case <-s.ctx.Done():
			s.Logger().Info("all module done")
			return
		case <-time.After(time.Second):
			s.Logger().Infof("tick, settings: %+v...", s.SafeSettings())
		}
	}
}

func (s *SimpleModule) Type() interface{} {
	return (**SimpleModule)(nil)
}

func main() {
	gomodule.RegisterDefaultModule(configcenter.CC)
	gomodule.Register(instance)
	gomodule.RegisterDefaultModules()
	gomodule.Serv().Run(context.Background())
	// gomodule.Wait()
}
