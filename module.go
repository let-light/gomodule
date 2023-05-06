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

var defaultmanager *Manager

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
	wg             sync.WaitGroup
	defaultModules []*ModuleInfo
	plogger        *logrus.Entry
	roomCmdRun     bool
	servctl        *servctl
}

type IModule interface {
	InitModule(ctx context.Context, m *Manager) (interface{}, error)
	InitCommand() ([]*cobra.Command, error)
	ConfigChanged()
	PreModuleRun()
	ModuleRun()
}

func init() {
	defaultmanager = NewManager()
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

	m.logger().Infof("receive signal: %v", sig)

	m.cancel()
}

func (m *Manager) logger() *logrus.Entry {
	return m.plogger
}

func NewManager() *Manager {
	m := &Manager{
		modules:        make([]*ModuleInfo, 0),
		rootCmd:        &cobra.Command{},
		defaultModules: make([]*ModuleInfo, 0),
		plogger:        logrus.WithField("module", "manager"),
		roomCmdRun:     false,
	}

	m.servctl = newServctl(m)

	m.rootCmd.Run = func(cmd *cobra.Command, args []string) {
		m.roomCmdRun = true
	}
	return m
}

func (m *Manager) initDefaultModules() {
	for _, dmi := range m.defaultModules {
		for _, mi := range m.modules {
			if mi.name == dmi.name {
				m.logger().Panic(fmt.Errorf("module[%s] is existed", dmi.name))
			}
		}
	}

	modules := m.defaultModules
	modules = append(modules, m.modules...)
	m.modules = modules
}

func (m *Manager) Register(module IModule) error {
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
	for _, mi := range m.modules {
		if mi.name == t.Elem().Name() {
			return fmt.Errorf("module[%+v] is existed", mi)
		}
	}

	m.logger().Debugf("get module named: %s", name)

	return m.RegisterWithName(module, name)
}

func (m *Manager) RegisterDefaultModule(module IModule) error {
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
	for _, mi := range m.modules {
		if mi.name == name {
			return fmt.Errorf("module[%+v] is existed", mi)
		}
	}

	m.logger().Debugf("get module named: %s", name)

	return m.RegisterDefaultModuleWithName(module, name)
}

func (m *Manager) RegisterDefaultModules() {
	if e := RegisterDefaultModuleWithName(ConfigModule(), "config"); e != nil {
		m.logger().Panic(e)
	}

	if e := RegisterDefaultModuleWithName(LoggerModule(), "logger"); e != nil {
		m.logger().Panic(e)
	}
}

func (m *Manager) RegisterWithName(module IModule, name string) error {
	t := reflect.TypeOf(module)
	if t.Kind() != reflect.Ptr {
		m.logger().Error("module must be pointer")
		return fmt.Errorf("module must be pointer")
	}

	if t.Elem().Kind() != reflect.Struct {
		m.logger().Error("module must be struct")
		return fmt.Errorf("module must be struct")
	}

	m.modules = append(m.modules, &ModuleInfo{
		module: module,
		cmds:   make([]*cobra.Command, 0),
		name:   name,
	})

	m.logger().Debugf("register module: %s", name)

	return nil
}

func (m *Manager) RegisterDefaultModuleWithName(module IModule, name string) error {
	t := reflect.TypeOf(module)
	if t.Kind() != reflect.Ptr {
		m.logger().Error("module must be pointer")
		return fmt.Errorf("module must be pointer")
	}

	if t.Elem().Kind() != reflect.Struct {
		m.logger().Error("module must be struct")
		return fmt.Errorf("module must be struct")
	}

	m.defaultModules = append(m.defaultModules, &ModuleInfo{
		module: module,
		cmds:   make([]*cobra.Command, 0),
		name:   name,
	})

	m.logger().Debugf("register default module: %s", name)

	return nil
}

func (m *Manager) Launch(ctx context.Context) error {
	if e := m.initModules(ctx); e != nil {
		return e
	}
	go m.sysSignal()

	m.logger().Debugf("launch modules")

	if e := m.execute(); e != nil {
		return e
	}

	if m.roomCmdRun {
		m.run()
	}

	return nil
}

func (m *Manager) Run(ctx context.Context) error {
	if e := m.Launch(ctx); e != nil {
		return e
	}

	m.Wait()

	return nil
}

func (m *Manager) GetRootCmd() *cobra.Command {
	return m.rootCmd
}

func (m *Manager) Wait() {
	go func() {
		m.wg.Wait()
		m.logger().Debug("all modules run done")
		m.cancel()
	}()

	<-m.ctx.Done()
}

func (m *Manager) configChanged() {
	for _, mi := range m.modules {
		mi.module.ConfigChanged()
	}
}

func (m *Manager) initWaitGroup() {
	m.once.Do(func() {
		m.logger().Debug("init wait group")
		m.wg.Add(len(m.modules))
	})
}

func (m *Manager) run() {
	for _, mi := range m.modules {
		m.logger().Debugf("pre module run: %s", mi.name)
		mi.module.PreModuleRun()
	}

	m.initWaitGroup()

	for _, mi := range m.modules {
		m.logger().Debugf("module run: %s", mi.name)
		go func(mi *ModuleInfo) {
			defer m.wg.Done()
			mi.module.ModuleRun()
		}(mi)
	}
}

func (m *Manager) execute() error {
	return m.rootCmd.Execute()
}

func (m *Manager) initModules(ctx context.Context) error {
	m.ctx, m.cancel = context.WithCancel(ctx)
	m.initDefaultModules()

	m.logger().Debug("modules: ", len(m.modules))
	m.logger().Debug("default modules: ", len(m.defaultModules))

	// init module
	for _, mi := range m.modules {
		settings, err := mi.module.InitModule(m.ctx, m)
		if err != nil {
			return err
		} else if settings == nil {
			m.logger().Debugf("module[%s] settings is nil", mi.name)
		}
		mi.settings = settings
	}

	m.logger().Debugf("launch defaultmanager, modules: %+v", m.modules)

	// init command
	for _, mi := range m.modules {
		cmds, err := mi.module.InitCommand()
		if err != nil {
			return err
		}

		for _, cmd := range cmds {
			m.rootCmd.AddCommand(cmd)
		}

		mi.cmds = cmds
	}

	return nil
}

func (m *Manager) Stop() {
	m.cancel()
}

func (m *Manager) Serv() *servctl {
	return m.servctl
}

func Register(module IModule) error {
	return defaultmanager.Register(module)
}

func RegisterDefaultModule(module IModule) error {
	return defaultmanager.RegisterDefaultModule(module)
}

func RegisterDefaultModules() {
	defaultmanager.RegisterDefaultModules()
}

func RegisterWithName(module IModule, name string) error {
	return defaultmanager.RegisterWithName(module, name)
}

func RegisterDefaultModuleWithName(module IModule, name string) error {
	return defaultmanager.RegisterDefaultModuleWithName(module, name)
}

func Launch(ctx context.Context) error {
	return defaultmanager.Launch(ctx)
}

func Run(ctx context.Context) error {
	return defaultmanager.Run(ctx)
}

func GetRootCmd() *cobra.Command {
	return defaultmanager.GetRootCmd()
}

func Wait() {
	defaultmanager.Wait()
}

func Stop() {
	defaultmanager.Stop()
}

func Serv() *servctl {
	return defaultmanager.Serv()
}
