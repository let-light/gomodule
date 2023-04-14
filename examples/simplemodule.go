package main

import (
	"context"
	"sync"
	"time"

	"github.com/let-light/gomodule"
	"github.com/let-light/gomodule/examples/configcenter"
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
	flags    *MainFlags
	settings *Settings
	wg       *sync.WaitGroup
	ctx      context.Context
	logger   *logrus.Entry
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

func (s *SimpleModule) InitModule(ctx context.Context, wg *sync.WaitGroup) (interface{}, error) {
	s.wg = wg
	s.ctx = ctx
	s.Logger().Info("init simple module")
	return s.settings, nil
}

func (s *SimpleModule) RootCommand(cmd *cobra.Command, args []string) {
	s.Logger().Info("root command")

	s.Logger().Info("settings: ", s.settings.Test)
	done := make(chan struct{})
	s.wg.Add(1)
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				s.Logger().Info("all module done")
				done <- struct{}{}
			case <-done:
				s.Logger().Info("simple module done")
				s.wg.Done()
				return
			case <-time.After(time.Second):
				s.Logger().Info("tick...")
			}
		}
	}()

}

func main() {
	gomodule.RegisterDefaultModule(configcenter.CC)
	gomodule.Register(instance)
	gomodule.RegisterDefaultModules()
	gomodule.Launch(context.Background())
	gomodule.Wait()
}
