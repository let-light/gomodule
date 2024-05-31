package configcenter

import (
	"context"

	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/let-light/gomodule"
	feature_configcenter "github.com/let-light/gomodule/examples/features/configcenter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type LoggerInfo struct {
	Formatter    string `json:"formatter"`
	Format       string `json:"format"`
	File         string `json:"file"`
	Console      bool   `json:"console"`
	Level        string `json:"level"`
	ReportCaller bool   `json:"reportCaller"`
}

type SimpleModuleInfo struct {
	Test string `json:"test"`
}

type RemoteConfig struct {
	Logger LoggerInfo       `json:"loggerModule"`
	Simple SimpleModuleInfo `json:"SimpleModule"`
}

type ConfigCenter struct {
	gomodule.DefaultModule
	ctx context.Context
	s   *ghttp.Server
}

var CC *ConfigCenter

func init() {
	CC = &ConfigCenter{}
}

func (c *ConfigCenter) InitModule(ctx context.Context, _ *gomodule.Manager) (interface{}, error) {
	c.ctx = ctx
	logrus.Info("init configcenter module")
	return nil, nil
}

func (c *ConfigCenter) InitCommand() ([]*cobra.Command, error) {
	logrus.Info("init command")
	return nil, nil
}

func (c *ConfigCenter) ConfigChanged() {

}

func (c *ConfigCenter) PreModuleRun() {
	logrus.Info("pre module run")
	s := g.Server()
	s.BindHandler("/", func(r *ghttp.Request) {
		r.Response.WriteJsonExit(&RemoteConfig{
			Logger: LoggerInfo{
				Formatter:    "text",
				Format:       "2006-01-02 15:04:05",
				File:         "logs/gomodule.log",
				Console:      true,
				Level:        "info",
				ReportCaller: true,
			},
			Simple: SimpleModuleInfo{
				Test: "test",
			},
		})
	})
	s.SetPort(9990)
	go func() {
		<-c.ctx.Done()
		logrus.Info("config center done")
		s.Shutdown()
	}()

	c.s = s
}

func (c *ConfigCenter) ModuleRun() {
	logrus.Info("module run")
	c.s.Run()
}

func (c *ConfigCenter) Type() interface{} {
	return (*feature_configcenter.Feature)(nil)
}

func (c *ConfigCenter) HelloWorld() {
	logrus.Info("hello World !")
}
