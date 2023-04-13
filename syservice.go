package gomodule

import (
	"context"
	"fmt"
	"os/user"
	"strings"
	"sync"

	"github.com/gogf/gf/os/gfile"
	"github.com/kardianos/service"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type sysFlags struct {
	Service     string `env:"service" flag:"service"`
	Display     string `env:"display" flag:"display"`
	Description string `env:"description"   flag:"description"`
	WorkDir     string `env:"workdir" flag:"workdir"`
	Args        string `env:"args" flag:"args"`
}

type syservice struct {
	ctx    context.Context
	cancel context.CancelFunc
	svc    service.Service
	flags  sysFlags
}

var serviceInstance *syservice

func init() {

	ctx, cancel := context.WithCancel(context.Background())
	serviceInstance = &syservice{
		ctx:    ctx,
		cancel: cancel,
	}
}

func SyServiceModule() *syservice {
	return serviceInstance
}

func (s *syservice) Start(ss service.Service) error {
	return nil
}

func (s *syservice) Stop(ss service.Service) error {
	logrus.Info("service stop")
	s.cancel()
	return nil
}

func (s *syservice) InitModule(ctx context.Context, wg *sync.WaitGroup) (interface{}, error) {
	return nil, nil
}

func (s *syservice) InitCommand() ([]*cobra.Command, error) {
	u, err := user.Current()
	if err != nil {
		fmt.Printf("get current user failed, err %v", err)
		return nil, nil
	}

	fmt.Printf("current user %+v\n", u)
	cmd := &cobra.Command{
		Use:   "service",
		Short: `[install|uninstall|start|stop]`,
		Long:  `it is a command for service install uninstall start stop`,
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			s.svc, err = service.New(s, &service.Config{
				Name:             s.flags.Service,
				DisplayName:      s.flags.Display,
				Description:      s.flags.Description,
				WorkingDirectory: s.flags.WorkDir,
				Arguments:        strings.Split(s.flags.Args, " "),
			})

			if err != nil {
				logrus.Error("new service failed:", err)
				return
			}

			install, err := cmd.Flags().GetBool("install")
			if err != nil {
				logrus.Error("service command get install flag error:", err)
				return
			}

			uninstall, err := cmd.Flags().GetBool("uninstall")
			if err != nil {
				logrus.Error("service command get uninstall flag error:", err)
				return
			}

			start, err := cmd.Flags().GetBool("start")
			if err != nil {
				logrus.Error("service command get start flag error:", err)
				return
			}

			stop, err := cmd.Flags().GetBool("stop")
			if err != nil {
				logrus.Error("service command get stop flag error:", err)
				return
			}

			restart, err := cmd.Flags().GetBool("restart")
			if err != nil {
				logrus.Error("service command get restart flag error:", err)
				return
			}

			if install {
				err = s.svc.Install()
				if err != nil {
					logrus.Errorf("service install error: ", err)
					return
				}
				logrus.Infof("service install success")
			}

			if uninstall {
				err = s.svc.Uninstall()
				if err != nil {
					logrus.Errorf("service uninstall error: ", err)
					return
				}
				logrus.Infof("service uninstall success")
				return
			}

			if start {
				err = s.svc.Start()
				if err != nil {
					logrus.Errorf("service start error: ", err)
					return
				}
				logrus.Infof("service start success")
			}

			if stop {
				err = s.svc.Stop()
				if err != nil {
					logrus.Errorf("service stop error: ", err)
					return
				}
				logrus.Infof("service stop success")
			}

			if restart {
				err = s.svc.Restart()
				if err != nil {
					logrus.Errorf("service restart error: ", err)
					return
				}
				logrus.Infof("service restart success")
			}
		},
	}

	cmd.Flags().BoolP("install", "i", false, "install your service")
	cmd.Flags().BoolP("uninstall", "u", false, "uninstall your service")
	cmd.Flags().BoolP("start", "s", false, "start your service")
	cmd.Flags().BoolP("stop", "t", false, "stop your service")
	cmd.Flags().BoolP("restart", "r", false, "restart your service")
	cmd.Flags().StringVar(&s.flags.Service, "service", "", "service name")
	cmd.Flags().StringVar(&s.flags.Display, "display", "", "display name")
	cmd.Flags().StringVar(&s.flags.Description, "description", "", "description")
	cmd.Flags().StringVar(&s.flags.WorkDir, "workdir", gfile.SelfDir(), "workdir")
	cmd.Flags().StringVar(&s.flags.Args, "args", "", "args")

	return []*cobra.Command{cmd}, nil

}

func (s *syservice) ConfigChanged() {

}

func (s *syservice) RootCommand(cmd *cobra.Command, args []string) {
	if s.svc == nil {
		return
	}

	go s.svc.Run()
}
