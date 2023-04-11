package gomodule

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var manager *Manager

type ModuleInfo struct {
	module   IModule
	settings interface{}
	cmds     []*cobra.Command
	name     string
}

type Manager struct {
	modules        []*ModuleInfo
	rootCmd        *cobra.Command
	once           sync.Once
	ctx            context.Context
	cancel         context.CancelFunc
	wg             *sync.WaitGroup
	defaultModules []IModule
}

type IModule interface {
	InitModule(ctx context.Context, wg *sync.WaitGroup) (interface{}, error)
	InitCommand() ([]*cobra.Command, error)
	ConfigChanged()
	RootCommand(cmd *cobra.Command, args []string)
}

func init() {
	manager = NewManager()
}

func (m *Manager) sysSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch,
		syscall.SIGHUP,
		syscall.SIGINT,  // Ctrl+C
		syscall.SIGQUIT, // Ctrl+\
		syscall.SIGILL,  // illegal instruction
		syscall.SIGABRT, // abort() called
		syscall.SIGFPE,  // floating point exception
		syscall.SIGSEGV, // segmentation violation
		syscall.SIGPIPE, // broken pipe
		syscall.SIGTERM, // software termination signal from kill
	)

	sig := <-ch

	logrus.Infof("receive signal: %v", sig)

	m.cancel()
}

func NewManager() *Manager {
	m := &Manager{
		modules:        make([]*ModuleInfo, 0),
		rootCmd:        &cobra.Command{},
		defaultModules: make([]IModule, 0),
	}

	m.wg = &sync.WaitGroup{}
	m.ctx, m.cancel = context.WithCancel(context.Background())

	m.rootCmd.Run = func(cmd *cobra.Command, args []string) {
		for _, mi := range manager.modules {
			mi.module.RootCommand(cmd, args)
		}
	}

	go m.sysSignal()

	return m
}

func initDefaultModules() {
	modules := make([]*ModuleInfo, 0)
	for _, module := range manager.defaultModules {
		if module == nil {
			fmt.Printf("module is nil")
			continue
		}

		t := reflect.TypeOf(module)
		if t.Kind() != reflect.Ptr {
			fmt.Printf("module must be pointer")
			continue
		}

		if t.Elem().Kind() != reflect.Struct {
			fmt.Printf("module must be struct")
			continue
		}

		for _, mi := range manager.modules {
			if mi.module == module {
				fmt.Printf("module[%p] is existed", module)
				return
			}
		}

		mi := &ModuleInfo{
			module: module,
			cmds:   make([]*cobra.Command, 0),
			name:   t.Elem().Name(),
		}

		modules = append(modules, mi)
	}

	manager.modules = append(modules, manager.modules...)
}

func Register(modules ...IModule) error {
	if len(modules) == 0 {
		return nil
	}

	for _, module := range modules {
		t := reflect.TypeOf(module)
		if t.Kind() != reflect.Ptr {
			return fmt.Errorf("module must be pointer")
		}

		if t.Elem().Kind() != reflect.Struct {
			return fmt.Errorf("module must be struct")
		}

		for _, mi := range manager.modules {
			if mi.module == module {
				return fmt.Errorf("module[%p] is existed", module)
			}
		}

		mi := &ModuleInfo{
			module: module,
			cmds:   make([]*cobra.Command, 0),
			name:   t.Elem().Name(),
		}

		manager.modules = append(manager.modules, mi)
	}

	return nil
}

func RegisterDefaultModule(modules ...IModule) {
	manager.defaultModules = append(manager.defaultModules, modules...)
}

func RegisterDefaultModules() {
	RegisterDefaultModule(ConfigModule(), LoggerModule())
}

func Launch(ctx context.Context) error {
	manager.ctx, manager.cancel = context.WithCancel(ctx)
	manager.once.Do(initDefaultModules)

	// init module
	for _, mi := range manager.modules {
		settings, err := mi.module.InitModule(ctx, manager.wg)
		if err != nil {
			return err
		}
		mi.settings = settings
	}

	// init command
	for _, mi := range manager.modules {
		cmds, err := mi.module.InitCommand()
		if err != nil {
			return err
		}

		for _, cmd := range cmds {
			manager.rootCmd.AddCommand(cmd)
		}

		mi.cmds = cmds
	}

	manager.rootCmd.Execute()

	return nil
}

func GetRootCmd() *cobra.Command {
	return manager.rootCmd
}

func Wait() {
	go func() {
		manager.wg.Wait()
		manager.cancel()
	}()

	<-manager.ctx.Done()
}

func ConfigChanged() {
	for _, mi := range manager.modules {
		mi.module.ConfigChanged()
	}
}
