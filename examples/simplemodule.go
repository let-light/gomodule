package main

import (
	"context"
	"sync"
	"time"

	"github.com/let-light/gomodule"
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
}

var instance *SimpleModule

func init() {
	instance = &SimpleModule{
		flags:    &MainFlags{},
		settings: &Settings{},
	}
	gomodule.Register(instance)
}

func (c *SimpleModule) InitModule(ctx context.Context, wg *sync.WaitGroup) (interface{}, error) {
	c.wg = wg
	c.wg.Add(1)
	logrus.Info("init simple module")
	return c.settings, nil
}

func (c *SimpleModule) InitCommand() ([]*cobra.Command, error) {
	logrus.Info("init command")
	return nil, nil
}

func (c *SimpleModule) ConfigChanged() {
	logrus.Info("config changed")

}

func (c *SimpleModule) RootCommand(cmd *cobra.Command, args []string) {
	logrus.Info("root command")

	logrus.Info("settings: ", c.settings.Test)

	go func() {
		logrus.Info("wait for 5 seconds ...")
		<-time.After(5 * time.Second)
		c.wg.Done()
	}()
}

func main() {
	gomodule.RegisterDefaultModules()
	gomodule.Register(instance)
	gomodule.Launch(context.Background())
	gomodule.Wait()
}
