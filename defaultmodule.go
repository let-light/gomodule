package gomodule

import (
	"context"

	"github.com/spf13/cobra"
)

type DefaultModule struct {
	Manager *Manager
}

func (dm *DefaultModule) InitModule(ctx context.Context, m *Manager) (interface{}, error) {
	return nil, nil
}

func (dm *DefaultModule) InitCommand() ([]*cobra.Command, error) {
	return nil, nil
}

func (dm *DefaultModule) ConfigChanged() {

}

func (dm *DefaultModule) PreModuleRun() {

}

func (dm *DefaultModule) ModuleRun() {

}
