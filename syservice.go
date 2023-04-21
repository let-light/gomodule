package gomodule

import (
	"context"
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
	DefaultModule
	ctx    context.Context
	cancel context.CancelFunc
	svc    service.Service
	flags  sysFlags
	logger *logrus.Entry
}

var serviceInstance *syservice

func init() {

	ctx, cancel := context.WithCancel(context.Background())
	serviceInstance = &syservice{
		ctx:    ctx,
		cancel: cancel,
		logger: logrus.WithField("module", "syserver"),
	}
}

func SyServiceModule() *syservice {
	return serviceInstance
}

func (s *syservice) Logger() *logrus.Entry {
	return s.logger
}

func (s *syservice) Start(ss service.Service) error {
	return nil
}

func (s *syservice) Stop(ss service.Service) error {
	s.Logger().Info("service stop")
	s.cancel()
	return nil
}

func (s *syservice) InitModule(ctx context.Context, wg *sync.WaitGroup) (interface{}, error) {
	s.Logger().Info("init syservice module")
	return nil, nil
}

func (s *syservice) InitCommand() ([]*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "service",
		Short: `Service manager, --help for more info`,
		Long:  `it is a command for service install uninstall start stop`,
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			s.svc, err = service.New(s, &service.Config{
				Name:             s.flags.Service,
				DisplayName:      s.flags.Display,
				Description:      s.flags.Description,
				WorkingDirectory: s.flags.WorkDir,
				Arguments:        strings.Split(s.flags.Args, " "),
				Executable:       gfile.SelfPath(),
			})

			if err != nil {
				s.Logger().Error("new service failed:", err)
				return
			}

			install, err := cmd.Flags().GetBool("install")
			if err != nil {
				s.Logger().Error("service command get install flag error:", err)
				return
			}

			uninstall, err := cmd.Flags().GetBool("uninstall")
			if err != nil {
				s.Logger().Error("service command get uninstall flag error:", err)
				return
			}

			start, err := cmd.Flags().GetBool("start")
			if err != nil {
				s.Logger().Error("service command get start flag error:", err)
				return
			}

			stop, err := cmd.Flags().GetBool("stop")
			if err != nil {
				s.Logger().Error("service command get stop flag error:", err)
				return
			}

			restart, err := cmd.Flags().GetBool("restart")
			if err != nil {
				s.Logger().Error("service command get restart flag error:", err)
				return
			}

			if install {
				err = s.svc.Install()
				if err != nil {
					s.Logger().Errorf("service install error: ", err)
					return
				}
				s.Logger().Infof("service install success")
			}

			if restart {
				err = s.svc.Restart()
				if err != nil {
					s.Logger().Errorf("service restart error: ", err)
					return
				}
				s.Logger().Infof("service restart success")
			}

			if start {
				err = s.svc.Start()
				if err != nil {
					s.Logger().Errorf("service start error: ", err)
					return
				}
				s.Logger().Infof("service start success")
			}

			if stop {
				err = s.svc.Stop()
				if err != nil {
					s.Logger().Errorf("service stop error: ", err)
					return
				}
				s.Logger().Infof("service stop success")
			}

			if uninstall {
				err = s.svc.Uninstall()
				if err != nil {
					s.Logger().Errorf("service uninstall error: ", err)
					return
				}
				s.Logger().Infof("service uninstall success")
				return
			}
		},
	}

	cmd.PersistentFlags().BoolP("install", "i", false, "install your service")
	cmd.PersistentFlags().BoolP("uninstall", "u", false, "uninstall your service")
	cmd.PersistentFlags().BoolP("start", "s", false, "start your service")
	cmd.PersistentFlags().BoolP("stop", "t", false, "stop your service")
	cmd.PersistentFlags().BoolP("restart", "r", false, "restart your service")
	cmd.PersistentFlags().StringVar(&s.flags.Service, "service", "", "service name for service")
	cmd.PersistentFlags().StringVar(&s.flags.Display, "display", "", "display name for service")
	cmd.PersistentFlags().StringVar(&s.flags.Description, "description", "", "description for service")
	cmd.PersistentFlags().StringVar(&s.flags.WorkDir, "workdir", gfile.Pwd(), "workdir path for service run")
	cmd.PersistentFlags().StringVar(&s.flags.Args, "args", "", "cmd args for service run, split by space(' ')")

	return []*cobra.Command{cmd}, nil

}

func (s *syservice) RootCommand(cmd *cobra.Command, args []string) {
	if s.svc == nil {
		return
	}

	go s.svc.Run()
}
