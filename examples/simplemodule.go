package main

import (
	"context"
	"sync"
	"time"

	"github.com/let-light/gomodule"
	"github.com/let-light/gomodule/examples/configcenter"
	_ "github.com/let-light/gomodule/examples/configcenter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type MainFlags struct {
}

type Settings struct {
	Test string `mapstructure:"test"`
}

type SimpleModule struct {
	flags    *MainFlags
	settings *Settings
	wg       *sync.WaitGroup
	ctx      context.Context
}

var instance *SimpleModule

func init() {
	instance = &SimpleModule{
		flags:    &MainFlags{},
		settings: &Settings{},
	}
}

func (c *SimpleModule) InitModule(ctx context.Context, wg *sync.WaitGroup) (interface{}, error) {
	c.wg = wg
	c.ctx = ctx
	logrus.Info("init simple module")
	return c.settings, nil
}

func (c *SimpleModule) InitCommand() ([]*cobra.Command, error) {
	logrus.Info("init command")
	return nil, nil
}

func (c *SimpleModule) ConfigChanged() {
	logrus.Info("config changed ", *c.settings)
}

func (c *SimpleModule) RootCommand(cmd *cobra.Command, args []string) {
	logrus.Info("root command")

	logrus.Info("settings: ", c.settings.Test)
	done := make(chan struct{})
	c.wg.Add(1)
	go func() {
		for {
			select {
			case <-c.ctx.Done():
				logrus.Info("all module done")
				done <- struct{}{}
			case <-done:
				logrus.Info("simple module done")
				c.wg.Done()
				return
			case <-time.After(time.Second):
				logrus.Info("tick...")
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
