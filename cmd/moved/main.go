package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/genzai-io/sliced/app/core"
	"github.com/genzai-io/sliced/common/service"
	"github.com/spf13/cobra"

	//_ "go.uber.org/automaxprocs"

	"github.com/genzai-io/sliced"
	_ "github.com/rs/zerolog/log"

	"github.com/genzai-io/sliced/app/server"
	"github.com/rs/zerolog"
)

func main() {
	var cmdStart = &cobra.Command{
		Use:   "start",
		Short: "Starts the app",
		Long:  ``,
		Args:  cobra.MinimumNArgs(0),
		Run:   start,
	}
	configureStart(cmdStart)

	var cmdStatus = &cobra.Command{
		Use:   "status",
		Short: "Prints the current status of " + moved.Name,
		Long:  ``,
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			l := moved.Lock()
			if l.Success {
				moved.Unlock()
				moved.Logger.Info().Msg("daemon is not running")
				return
			} else {
				moved.Logger.Info().Msgf("daemon pid %d", l.Pid)
			}
		},
	}

	var force bool
	var cmdStop = &cobra.Command{
		Use:   "stop",
		Short: "Stops the daemon process if it's running",
		Long:  `Determines the daemon PID from the daemon pid lock file and sends a SIGTERM signal if running.`,
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			stop(force)
		},
	}
	cmdStop.Flags().BoolVarP(
		&force,
		"force",
		"f",
		false,
		"Force stop the daemon process. KILL the process.",
	)

	var cmdRoot = &cobra.Command{
		Use: moved.Name,
		// Default to start as daemon
		Run: start,
	}
	configureStart(cmdRoot)
	cmdRoot.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Long:  `All software has versions. This is ` + moved.Name + `'s`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(fmt.Sprintf("%s %s", moved.Name, moved.VersionStr))
		},
	})
	cmdRoot.AddCommand(cmdStart, cmdStop, cmdStatus)

	//moved.BindCLI(cmdRoot)
	//moved.BindCLI(cmdStart)

	// Let's get started!
	cmdRoot.Execute()
}

func configureStart(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(
		&moved.Console,
		"console",
		"c",
		false,
		"Fmt logger for the console",
	)
	cmd.Flags().BoolVarP(
		&moved.Bootstrap,
		"bootstrap",
		"b",
		false,
		"Bootstrap cluster",
	)
	cmd.Flags().StringVarP(
		&moved.ApiHost,
		"api.host",
		"a",
		moved.ApiHost,
		"Host to run the API server on",
	)
	cmd.Flags().StringVarP(
		&moved.WebHost,
		"web.host",
		"w",
		moved.WebHost,
		"Host to run the WEB server on",
	)
	cmd.Flags().StringVarP(
		&moved.RaftHost,
		"raft.host",
		"r",
		moved.RaftHost,
		"Host to run the Raft cluster transport on",
	)
	cmd.Flags().IntVarP(
		&moved.EventLoops,
		"loops",
		"l",
		moved.EventLoops,
		"Number of API event loops to run",
	)
	cmd.Flags().IntVarP(
		&moved.LogLevel,
		"log",
		"v",
		int(zerolog.DebugLevel),
		"Log level",
	)
	cmd.Flags().StringVarP(
		&moved.DataDir,
		"pid",
		"p",
		"0",
		"Name of PID file",
	)
	cmd.Flags().StringVarP(
		&moved.DataDir,
		"data.path",
		"d",
		"",
		"Path to persist slice data",
	)
	cmd.Flags().StringVarP(
		&moved.StoreDir,
		"store.path",
		"s",
		moved.StoreDir,
		"Path where stores should persist to. Use \":memory:\" to disable disk persistence",
	)
	//cmd.Flags().BoolVarP(
	//	&singlemode,
	//	"singlemode",
	//	"s",
	//	false,
	//	"Start the cluster in single mode",
	//)
	moved.BindCLI(cmd)
}

func start(cmd *cobra.Command, args []string) {
	moved.Configure()

	if !moved.Console {
		// Ensure only 1 instance through PID lock
		lockResult := moved.Lock()
		if !lockResult.Success {
			moved.Logger.Error().Msgf("process already running pid:%d -- localhost:%d", lockResult.Pid, lockResult.Port)
			return
		}
		defer moved.Unlock()
	}

	// Create, Start and Wait for Daemon to exit
	app := &Daemon{}
	app.BaseService = *service.NewBaseService(moved.Logger, "daemon", app)
	err := app.Start()
	if err != nil {
		moved.Logger.Error().Err(err)
		return
	}
	app.Wait()
}

func stop(force bool) {
	// Ensure only 1 instance.
	l := moved.Lock()
	if l.Success {
		moved.Unlock()
		moved.Logger.Info().Msg("daemon is not running")
		return
	}

	if l.Pid > 0 {
		process, err := os.FindProcess(l.Pid)
		if err != nil {
			moved.Logger.Info().Msgf("failed to find daemon pid %d", l.Pid)
			moved.Logger.Error().Err(err)
		} else if process == nil {
			moved.Logger.Info().Msgf("failed to find daemon pid %d", l.Pid)
		} else {
			if force {
				moved.Logger.Info().Msgf("killing daemon pid %d", l.Pid)
				err = process.Kill()
				if err != nil {
					moved.Logger.Error().Err(err)
				} else {
					moved.Logger.Info().Msg("daemon was killed")
				}
			} else {
				moved.Logger.Info().Msgf("sending SIGTERM signal to pid %d", l.Pid)
				err := process.Signal(syscall.SIGTERM)
				if err != nil {
					moved.Logger.Error().Err(err)
				} else {
					moved.Logger.Info().Msgf("SIGTERM pid %d", l.Pid)
				}
			}
		}
	} else {
		moved.Logger.Info().Msg("daemon is not running")
	}
}

type Daemon struct {
	service.BaseService

	webServer *server.Web
	server    *server.CmdServer
}

func (d *Daemon) OnStart() error {
	//go metrics.Log(metrics.DefaultRegistry, 5 * time.Second, log.New(os.Stderr, "metrics: ", log.Lmicroseconds))

	// Handle os signals.
	c := make(chan os.Signal, 1)
	slogger := moved.Logger.With().Str("logger", "os.signal").Logger()
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		s := <-c
		// Log signal
		slogger.Info().Msgf("%s", s)

		// Stop app
		d.Stop()

		switch {
		default:
			os.Exit(-1)
		case s == syscall.SIGHUP:
			os.Exit(1)
		case s == syscall.SIGINT:
			os.Exit(2)
		case s == syscall.SIGQUIT:
			os.Exit(3)
		case s == syscall.SIGTERM:
			os.Exit(0xf)
		}
	}()

	if err := core.Instance.Start(); err != nil {
		return err
	}

	// Start Web.
	d.server = server.NewCmdServer()
	if err := d.server.Start(); err != nil {
		core.Instance.Stop()
		return err
	}

	d.webServer = server.NewWeb(moved.WebHost)
	if err := d.webServer.Start(); err != nil {
		d.server.Stop()
		core.Instance.Stop()
		return err
	}

	return nil
}

func (d *Daemon) OnStop() {
	if err := d.webServer.Stop(); err != nil {
		d.Logger.Error().Err(err)
	}
	if d.server != nil {
		if err := d.server.Stop(); err != nil {
			d.Logger.Error().Err(err)
		}
	}
	if err := core.Instance.Stop(); err != nil {
		d.Logger.Error().Err(err)
	}
}
