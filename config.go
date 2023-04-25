package gomodule

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gogf/gf/os/gfile"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configInstance *configModule

type ConfigSettings struct {
}

type configFlags struct {
	LocalFile          string `env:"localcfg" flag:"localcfg"`
	Consul             string `env:"consulcfg" flag:"consulcfg"`
	Etcd               string `env:"etcdcfg"   flag:"etcdcfg"`
	RemoteFile         string `env:"remotecfg" flag:"remotecfg"`
	RemoteFileInterval int    `env:"remotecfginterval" flag:"remotecfginterval"`
}

type configModule struct {
	DefaultModule
	flags    configFlags
	config   *viper.Viper
	settings ConfigSettings
	ctx      context.Context
	mtx      sync.Mutex
	logger   *logrus.Entry
	m        *Manager
}

func init() {
	configInstance = &configModule{
		logger: logrus.WithField("module", "config"),
	}
}

func ConfigModule() *configModule {
	return configInstance
}

func (c *configModule) Logger() *logrus.Entry {
	return configInstance.logger
}

func (c *configModule) Viper() *viper.Viper {
	return c.config
}

func (c *configModule) InitModule(ctx context.Context, m *Manager) (interface{}, error) {
	c.ctx = ctx
	c.m = m
	return &c.settings, nil
}

func (c *configModule) InitCommand() ([]*cobra.Command, error) {
	c.Logger().Debug("init config module")
	GetRootCmd().PersistentFlags().StringVarP(&c.flags.LocalFile, "cfg.local", "c", "", "Load config file")
	GetRootCmd().PersistentFlags().StringVar(&c.flags.Consul, "cfg.consul", "", "Load config file from consul")
	GetRootCmd().PersistentFlags().StringVar(&c.flags.Etcd, "cfg.etcd", "", "Load config file from etcd")
	GetRootCmd().PersistentFlags().StringVar(&c.flags.RemoteFile, "cfg.remote", "", "Load config file from remote api")
	GetRootCmd().PersistentFlags().IntVar(&c.flags.RemoteFileInterval, "cfg.remote.interval", 30, "Interval to reload config file from remote api")

	return nil, nil
}

func (c *configModule) loadConfigFromLocal() {
	if c.flags.LocalFile != "" {
		if !gfile.Exists(c.flags.LocalFile) {
			panic(fmt.Errorf("config file not found: %s", c.flags.LocalFile))
		} else {
			c.Logger().Infof("config file: %s\n", c.flags.LocalFile)
		}

		c.config.AddConfigPath(gfile.Dir(c.flags.LocalFile))
		c.config.SetConfigName(gfile.Basename(c.flags.LocalFile))
		c.config.SetConfigType(gfile.Ext(c.flags.LocalFile)[1:])
	} else {
		c.Logger().Info("default config file: config.yml (./config.yml or ./config/config.yml)")

		pwd, err := os.Getwd()
		if err != nil {
			panic(fmt.Errorf("get current work dir failed, %s", err))
		}

		c.config.AddConfigPath(pwd)
		c.config.AddConfigPath(".")        // optionally look for config in the working directory
		c.config.AddConfigPath("./config") // optionally look for config in the working directory
		c.config.SetConfigName("config")   // name of config file (without extension)
		c.config.SetConfigType("yml")      // REQUIRED if the config file does not have the extension in the name
	}

	err := c.config.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file, %s", err))
	}
	c.reloadSettings()

	c.config.OnConfigChange(func(e fsnotify.Event) {
		if err := c.config.ReadInConfig(); err != nil {
			panic(fmt.Errorf("fatal error config file, %s", err))
		}

		c.Logger().Debug("Config file changed:", e.Name)
		c.reloadSettings()

		c.m.configChanged()
	})
	c.config.WatchConfig()
}

func (c *configModule) loadConfigFromEtcd() {
	if c.flags.Etcd != "" {
		u, err := url.Parse(c.flags.Etcd)
		if err != nil {
			panic(fmt.Errorf("etcd url parse error, %s", err))
		}

		addr := u.Scheme + "://" + u.Host
		path := u.Path
		query := u.Query()

		fileType := ""
		ty := query["type"]
		if len(ty) > 0 {
			fileType = ty[0]
		}

		c.config.AddRemoteProvider("etcd", addr, path)
		c.config.SetConfigType(fileType)
		c.config.ReadRemoteConfig()
	}
}

func (c *configModule) loadConfigFromConsul() {
	if c.flags.Consul != "" {
		u, err := url.Parse(c.flags.Consul)
		if err != nil {
			panic(fmt.Errorf("consul url parse error, %s", err))
		}

		addr := u.Scheme + "://" + u.Host
		path := u.Path
		query := u.Query()

		fileType := ""
		ty := query["type"]
		if len(ty) > 0 {
			fileType = ty[0]
		}

		c.config.AddRemoteProvider("consul", addr, path)
		c.config.SetConfigType(fileType)
		c.config.ReadRemoteConfig()
	}
}

func (c *configModule) getRemoteFileContent() ([]byte, string, error) {
	resp, err := http.Get(c.flags.RemoteFile)
	if err != nil {
		// handle error
		return nil, "", fmt.Errorf("get config error, %s", err)
	}

	defer resp.Body.Close()

	configData, err := io.ReadAll(resp.Body)
	if err != nil {
		// handle error
		return nil, "", fmt.Errorf("read config error, %s", err)
	}

	ct := resp.Header.Get("Content-Type")

	if idx := strings.Index(ct, ";"); idx > 0 {
		ct = ct[:idx]
	}

	// "yaml", "yml", "json", "toml", "hcl", "tfvars", "ini", "prop", "props", "properties", "dotenv", "env"
	if strings.Contains(ct, "yaml") || strings.Contains(ct, "yml") {
		return configData, "yaml", nil
	} else if strings.Contains(ct, "toml") {
		return configData, "toml", nil
	} else if strings.Contains(ct, "json") {
		return configData, "json", nil
	} else if strings.Contains(ct, "prop") {
		return configData, "prop", nil
	} else if strings.Contains(ct, "props") {
		return configData, "props", nil
	} else if strings.Contains(ct, "tfvars") {
		return configData, "tfvars", nil
	} else if strings.Contains(ct, "properties") {
		return configData, "properties", nil
	} else if strings.Contains(ct, "hcl") {
		return configData, "hcl", nil
	} else if strings.Contains(ct, "ini") {
		return configData, "ini", nil
	} else if strings.Contains(ct, "dotenv") {
		return configData, "dotenv", nil
	} else if strings.Contains(ct, "env") {
		return configData, "env", nil
	} else {
		return nil, "", fmt.Errorf("unknown config type")
	}
}

func (c *configModule) loadConfigFromRemoteFile() {
	if c.flags.RemoteFile == "" {
		return
	}

	configData, ty, err := c.getRemoteFileContent()
	if err != nil {
		panic(fmt.Errorf("get config error, %s", err))
	}

	c.config.SetConfigType(ty)

	if e := c.config.ReadConfig(bytes.NewBuffer(configData)); e != nil {
		// handle error
		panic(fmt.Errorf("read config error, %s", e))
	}

	if e := c.reloadSettings(); e != nil {
		panic(fmt.Errorf("reload settings error, %s", e))
	}

	go func() {
		defer func() {
			if err := recover(); err != nil {
				c.Logger().Error("loadConfigFromRemoteFile error:", err)
			}
		}()

		for {
			select {
			case <-c.ctx.Done():
				return
			case <-time.After(time.Duration(c.flags.RemoteFileInterval) * time.Second):
				c.reloadSettings()
			}
		}
	}()
}

func (c *configModule) PreModuleRun() {
	c.config = viper.New()

	if c.flags.RemoteFile != "" {
		c.loadConfigFromRemoteFile()
	} else {
		c.loadConfigFromLocal()
		c.loadConfigFromEtcd()
		c.loadConfigFromConsul()
	}
}

func (c *configModule) reloadSettings() error {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	c.Logger().Debug("reload settings, modules:", len(c.m.modules))
	for _, mi := range c.m.modules {
		if mi.settings == nil {
			continue
		}
		c.Logger().Debug("reload settings:", mi.name)
		if err := c.config.UnmarshalKey(mi.name, mi.settings); err != nil {
			return fmt.Errorf("unmarshal config error, %s", err)
		}
	}

	c.m.configChanged()

	return nil
}
