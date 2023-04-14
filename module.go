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
	defaultModules []*ModuleInfo
	logger         *logrus.Entry
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

type DefaultModule struct {
}

func (dm *DefaultModule) InitModule(ctx context.Context, wg *sync.WaitGroup) (interface{}, error) {
	return nil, nil
}

func (dm *DefaultModule) InitCommand() ([]*cobra.Command, error) {
	return nil, nil
}

func (dm *DefaultModule) ConfigChanged() {

}

func (dm *DefaultModule) RootCommand(cmd *cobra.Command, args []string) {

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

	logger().Infof("receive signal: %v", sig)

	m.cancel()
}

func logger() *logrus.Entry {
	return manager.logger
}

func NewManager() *Manager {
	m := &Manager{
		modules:        make([]*ModuleInfo, 0),
		rootCmd:        &cobra.Command{},
		defaultModules: make([]*ModuleInfo, 0),
		logger:         logrus.WithField("module", "manager"),
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
	for _, dmi := range manager.defaultModules {
		for _, mi := range manager.modules {
			if mi.name == dmi.name {
				logger().Panic("module[%+v] is existed\n", dmi)
			}
		}
	}

	modules := manager.defaultModules
	modules = append(modules, manager.modules...)
	manager.modules = modules
}

func Register(module IModule) error {
	if module == nil {
		return fmt.Errorf("module is nil")
	}

	t := reflect.TypeOf(module)
	if t.Kind() != reflect.Ptr {
		return fmt.Errorf("module must be pointer")
	}

	if t.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("module must be struct")
	}

	name := t.Elem().Name()
	for _, mi := range manager.modules {
		if mi.name == t.Elem().Name() {
			return fmt.Errorf("module[%+v] is existed", mi)
		}
	}

	logger().Infof("get module named: %s", name)

	return RegisterWithName(module, name)
}

func RegisterDefaultModule(module IModule) error {
	if module == nil {
		return fmt.Errorf("module is nil")
	}

	t := reflect.TypeOf(module)
	if t.Kind() != reflect.Ptr {
		return fmt.Errorf("module must be pointer")
	}

	if t.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("module must be struct")
	}

	name := t.Elem().Name()
	for _, mi := range manager.modules {
		if mi.name == name {
			return fmt.Errorf("module[%+v] is existed", mi)
		}
	}

	logger().Infof("get module named: %s", name)

	return RegisterDefaultModuleWithName(module, name)
}

func RegisterDefaultModules() {
	if e := RegisterDefaultModuleWithName(SyServiceModule(), "syservice"); e != nil {
		logger().Panic(e)
	}

	if e := RegisterDefaultModuleWithName(ConfigModule(), "config"); e != nil {
		logger().Panic(e)
	}

	if e := RegisterDefaultModuleWithName(LoggerModule(), "logger"); e != nil {
		logger().Panic(e)
	}
}

func RegisterWithName(module IModule, name string) error {
	t := reflect.TypeOf(module)
	if t.Kind() != reflect.Ptr {
		logger().Info("module must be pointer")
		return fmt.Errorf("module must be pointer")
	}

	if t.Elem().Kind() != reflect.Struct {
		logger().Info("module must be struct")
		return fmt.Errorf("module must be struct")
	}

	manager.modules = append(manager.modules, &ModuleInfo{
		module: module,
		cmds:   make([]*cobra.Command, 0),
		name:   name,
	})

	logger().Infof("register module: %s", name)

	return nil
}

func RegisterDefaultModuleWithName(module IModule, name string) error {
	t := reflect.TypeOf(module)
	if t.Kind() != reflect.Ptr {
		logger().Info("module must be pointer")
		return fmt.Errorf("module must be pointer")
	}

	if t.Elem().Kind() != reflect.Struct {
		logger().Info("module must be struct")
		return fmt.Errorf("module must be struct")
	}

	manager.defaultModules = append(manager.defaultModules, &ModuleInfo{
		module: module,
		cmds:   make([]*cobra.Command, 0),
		name:   name,
	})

	logger().Infof("register default module: %s", name)

	return nil
}

func Launch(ctx context.Context) error {
	manager.ctx, manager.cancel = context.WithCancel(ctx)
	manager.once.Do(initDefaultModules)

	logger().Info("launch manager, modules: ", len(manager.modules))
	logger().Info("launch manager, default modules: ", len(manager.defaultModules))

	// init module
	for _, mi := range manager.modules {
		settings, err := mi.module.InitModule(ctx, manager.wg)
		if err != nil {
			return err
		} else if settings == nil {
			logger().Debugf("module[%s] settings is nil", mi.name)
		}
		mi.settings = settings
	}

	logger().Debugf("launch manager, modules: %+v", manager.modules)

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
