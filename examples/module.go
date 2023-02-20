package main

import (
	"context"
	"sync"

	"github.com/let-light/gomodule"
	"github.com/spf13/cobra"
)

type MainFlags struct {
}

type MainSettings struct {
}

type MainModule struct {
	flags    *MainFlags
	settings *MainSettings
	wg       *sync.WaitGroup
}

var instance *MainModule

func init() {
	instance = &MainModule{
		flags:    &MainFlags{},
		settings: &MainSettings{},
	}
	gomodule.Register(instance)
}

func MainModuleInstance() gomodule.IModule {
	return instance
}

func (c *MainModule) OnInitModule(ctx context.Context, wg *sync.WaitGroup) (interface{}, error) {
	c.wg = wg
	return c.settings, nil
}

func (c *MainModule) OnInitCommand() ([]*cobra.Command, error) {
	return nil, nil
}

func (c *MainModule) OnConfigModified() {

}

func (c *MainModule) OnPostInitCommand() {

}

func (c *MainModule) OnMainRun(cmd *cobra.Command, args []string) {
}

func main() {
	gomodule.Launch(context.Background())
	gomodule.Wait()
}
