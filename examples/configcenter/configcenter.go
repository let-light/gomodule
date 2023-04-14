package configcenter

import (
	"context"
	"sync"
	"time"

	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/let-light/gomodule"
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
	wg  *sync.WaitGroup
	ctx context.Context
}

var CC *ConfigCenter

func init() {
	CC = &ConfigCenter{}
}

func (c *ConfigCenter) InitModule(ctx context.Context, wg *sync.WaitGroup) (interface{}, error) {
	c.wg = wg
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

func (c *ConfigCenter) RootCommand(cmd *cobra.Command, args []string) {
	c.wg.Add(1)
	logrus.Info("root command")
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
		s.Run()
		c.wg.Done()
		logrus.Info("config center server done")
	}()

	go func() {
		<-c.ctx.Done()
		logrus.Info("config center done")
		s.Shutdown()
	}()

	// Wait for the HTTP server to start.
	time.Sleep(time.Second * 1)
}
