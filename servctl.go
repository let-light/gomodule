package gomodule

import (
	"context"
	"os"

	"github.com/gogf/gf/os/gfile"
	"github.com/kardianos/service"
	"github.com/spf13/cobra"
)

type servFlags struct {
	ctrl, name, display, desc, workdir, args string
}

type servctl struct {
	m      *Manager
	flags  servFlags
	cmdRun bool
	ctx    context.Context
	cancel context.CancelFunc
}

func newServctl(m *Manager) *servctl {
	return &servctl{
		m: m,
	}
}

func (s *servctl) Start(ss service.Service) error {
	s.m.logger().Info("start service")
	s.m.run()
	return nil
}

func (s *servctl) Stop(ss service.Service) error {
	s.cancel()
	return nil
}

func (s *servctl) Run(ctx context.Context) {
	s.ctx, s.cancel = context.WithCancel(ctx)

	cmd := &cobra.Command{
		Use:   "serv",
		Short: "service control",
		Run: func(cmd *cobra.Command, _ []string) {
			s.cmdRun = true
			s.m.logger().Debug("run as service")
			args := []string{
				"--serv.name=" + s.flags.name,
				"--serv.workdir=" + s.flags.workdir,
				"--serv.display=" + s.flags.display,
				"--serv.desc=" + s.flags.desc,
			}

			if s.flags.args != "" {
				args = append(args, s.flags.args)
			}

			svc, err := service.New(s, &service.Config{
				Name:             s.flags.name,
				DisplayName:      s.flags.display,
				Description:      s.flags.desc,
				WorkingDirectory: s.flags.workdir,
				Arguments:        args,
				Executable:       gfile.SelfPath(),
			})

			if err != nil {
				s.m.logger().Panic(err)
			}

			if s.flags.ctrl == "program" {
				s.m.initWaitGroup()
				go svc.Run()
				return
			}

			s.m.logger().Debugf("control service: %s", s.flags.ctrl)
			if e := service.Control(svc, s.flags.ctrl); e != nil {
				s.m.logger().WithError(e).Error("control service failed")
				s.m.logger().Panic(e)
			}
		},
	}

	cmd.PersistentFlags().StringVar(&s.flags.ctrl, "ctrl", "", "service control, start|stop|restart|install|uninstall")
	cmd.PersistentFlags().StringVar(&s.flags.name, "name", "", "service name, unique in system")
	cmd.PersistentFlags().StringVar(&s.flags.display, "display", "", "service display name")
	cmd.PersistentFlags().StringVar(&s.flags.desc, "desc", "", "service description")
	cmd.PersistentFlags().StringVar(&s.flags.workdir, "workdir", gfile.Pwd(), "service workdir, default is current dir")
	cmd.PersistentFlags().StringVar(&s.flags.args, "args", "", "service args")
	s.m.GetRootCmd().AddCommand(cmd)

	s.m.GetRootCmd().PersistentFlags().StringVar(&s.flags.name, "serv.name", "", "service name, don't use in program mode")
	s.m.GetRootCmd().PersistentFlags().StringVar(&s.flags.display, "serv.display", "", "service display name, don't use in program mode")
	s.m.GetRootCmd().PersistentFlags().StringVar(&s.flags.desc, "serv.desc", "", "service description, don't use in program mode")
	s.m.GetRootCmd().PersistentFlags().StringVar(&s.flags.workdir, "serv.workdir", "", "service workdir, don't use in program mode")

	if e := s.m.initModules(s.ctx); e != nil {
		s.m.logger().Panic(e)
	}

	go s.m.sysSignal()

	if e := s.m.execute(); e != nil {
		s.m.logger().Panic(e)
	}

	if !s.m.roomCmdRun {
		return
	}

	if !s.cmdRun && s.flags.name != "" {
		svc, err := service.New(s, &service.Config{
			Name:             s.flags.name,
			DisplayName:      s.flags.display,
			Description:      s.flags.desc,
			WorkingDirectory: s.flags.workdir,
		})

		if err != nil {
			s.m.logger().Panic(err)
		}
		os.Chdir(s.flags.workdir)

		s.m.logger().Debug("run as service")
		svc.Run()
		s.m.logger().Debug("service exit")
	} else {
		s.m.logger().Debug("run as program")
		s.m.run()
		s.m.Wait()
	}
}
