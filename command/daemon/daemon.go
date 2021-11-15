// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package daemon

import (
	"context"
	"time"

	"github.com/drone-runners/drone-runner-docker/engine"
	"github.com/drone-runners/drone-runner-docker/engine/compiler"
	"github.com/drone-runners/drone-runner-docker/engine/linter"
	"github.com/drone-runners/drone-runner-docker/engine/resource"
	"github.com/drone-runners/drone-runner-docker/internal/match"

	"github.com/drone/runner-go/client"
	"github.com/drone/runner-go/environ/provider"
	"github.com/drone/runner-go/handler/router"
	"github.com/drone/runner-go/logger"
	loghistory "github.com/drone/runner-go/logger/history"
	"github.com/drone/runner-go/pipeline/reporter/history"
	"github.com/drone/runner-go/pipeline/reporter/remote"
	"github.com/drone/runner-go/pipeline/runtime"
	"github.com/drone/runner-go/pipeline/uploader"
	"github.com/drone/runner-go/poller"
	"github.com/drone/runner-go/registry"
	"github.com/drone/runner-go/secret"
	"github.com/drone/runner-go/server"
	"github.com/drone/signal"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"gopkg.in/alecthomas/kingpin.v2"
)

// empty context.
var nocontext = context.Background()

type daemonCommand struct {
	envfile string
}

func (c *daemonCommand) run(*kingpin.ParseContext) error {
	// load environment variables from file.
	godotenv.Load(c.envfile)

	// load the configuration from the environment
	config, err := fromEnviron()
	if err != nil {
		return err
	}

	// setup the global logrus logger.
	setupLogger(config)

	ctx, cancel := context.WithCancel(nocontext)
	defer cancel()

	// listen for termination signals to gracefully shutdown
	// the runner daemon.
	ctx = signal.WithContextFunc(ctx, func() {
		println("received signal, terminating process")
		cancel()
	})

	cli := client.New(
		config.Client.Address,
		config.Client.Secret,
		config.Client.SkipVerify,
	)
	if config.Client.Dump {
		cli.Dumper = logger.StandardDumper(
			config.Client.DumpBody,
		)
	}
	cli.Logger = logger.Logrus(
		logrus.NewEntry(
			logrus.StandardLogger(),
		),
	)

	opts := engine.Opts{
		HidePull: !config.Docker.Stream,
	}
	engine, err := engine.NewEnv(opts)
	if err != nil {
		logrus.WithError(err).
			Fatalln("cannot load the docker engine")
	}
	for {
		err := engine.Ping(ctx)
		if err == context.Canceled {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err != nil {
			logrus.WithError(err).
				Errorln("cannot ping the docker daemon")
			time.Sleep(time.Second)
		} else {
			logrus.Debugln("successfully pinged the docker daemon")
			break
		}
	}

	remote := remote.New(cli)
	upload := uploader.New(cli)
	tracer := history.New(remote)
	hook := loghistory.New()
	logrus.AddHook(hook)

	runner := &runtime.Runner{
		Client:   cli,
		Machine:  config.Runner.Name,
		Environ:  config.Runner.Environ,
		Reporter: tracer,
		Lookup:   resource.Lookup,
		Lint:     linter.New().Lint,
		Match: match.Func(
			config.Limit.Repos,
			config.Limit.Events,
			config.Limit.Trusted,
		),
		Compiler: &compiler.Compiler{
			Clone:          config.Runner.Clone,
			Privileged:     append(config.Runner.Privileged, compiler.Privileged...),
			Networks:       config.Runner.Networks,
			NetworkOpts:    config.Runner.NetworkOpts,
			NetrcCloneOnly: config.Netrc.CloneOnly,
			Volumes:        config.Runner.Volumes,
			Resources: compiler.Resources{
				Memory:     config.Resources.Memory,
				MemorySwap: config.Resources.MemorySwap,
				CPUQuota:   config.Resources.CPUQuota,
				CPUPeriod:  config.Resources.CPUPeriod,
				CPUShares:  config.Resources.CPUShares,
				CPUSet:     config.Resources.CPUSet,
				ShmSize:    config.Resources.ShmSize,
			},
			Tmate: compiler.Tmate{
				Image:          config.Tmate.Image,
				Enabled:        config.Tmate.Enabled,
				Server:         config.Tmate.Server,
				Port:           config.Tmate.Port,
				RSA:            config.Tmate.RSA,
				ED25519:        config.Tmate.ED25519,
				AuthorizedKeys: config.Tmate.AuthorizedKeys,
			},
			Environ: provider.Combine(
				provider.Static(config.Runner.Environ),
				provider.External(
					config.Environ.Endpoint,
					config.Environ.Token,
					config.Environ.SkipVerify,
				),
			),
			Registry: registry.Combine(
				registry.File(
					config.Docker.Config,
				),
				registry.External(
					config.Registry.Endpoint,
					config.Registry.Token,
					config.Registry.SkipVerify,
				),
			),
			Secret: secret.Combine(
				secret.StaticVars(
					config.Runner.Secrets,
				),
				secret.External(
					config.Secret.Endpoint,
					config.Secret.Token,
					config.Secret.SkipVerify,
				),
			),
		},
		Exec: runtime.NewExecer(
			tracer,
			remote,
			upload,
			engine,
			config.Runner.Procs,
		).Exec,
	}

	poller := &poller.Poller{
		Client:   cli,
		Dispatch: runner.Run,
		Filter: &client.Filter{
			Kind:    resource.Kind,
			Type:    resource.Type,
			OS:      config.Platform.OS,
			Arch:    config.Platform.Arch,
			Variant: config.Platform.Variant,
			Kernel:  config.Platform.Kernel,
			Labels:  config.Runner.Labels,
		},
	}

	var g errgroup.Group
	server := server.Server{
		Addr: config.Server.Port,
		Handler: router.New(tracer, hook, router.Config{
			Username: config.Dashboard.Username,
			Password: config.Dashboard.Password,
			Realm:    config.Dashboard.Realm,
		}),
	}

	logrus.WithField("addr", config.Server.Port).
		Infoln("starting the server")

	g.Go(func() error {
		return server.ListenAndServe(ctx)
	})

	// Ping the server and block until a successful connection
	// to the server has been established.
	for {
		err := cli.Ping(ctx, config.Runner.Name)
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		if ctx.Err() != nil {
			break
		}
		if err != nil {
			logrus.WithError(err).
				Errorln("cannot ping the remote server")
			time.Sleep(time.Second)
		} else {
			logrus.Infoln("successfully pinged the remote server")
			break
		}
	}

	g.Go(func() error {
		logrus.WithField("capacity", config.Runner.Capacity).
			WithField("endpoint", config.Client.Address).
			WithField("kind", resource.Kind).
			WithField("type", resource.Type).
			WithField("os", config.Platform.OS).
			WithField("arch", config.Platform.Arch).
			Infoln("polling the remote server")

		poller.Poll(ctx, config.Runner.Capacity)
		return nil
	})

	err = g.Wait()
	if err != nil {
		logrus.WithError(err).
			Errorln("shutting down the server")
	}
	return err
}

// helper function configures the global logger from
// the loaded configuration.
func setupLogger(config Config) {
	logger.Default = logger.Logrus(
		logrus.NewEntry(
			logrus.StandardLogger(),
		),
	)
	if config.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	if config.Trace {
		logrus.SetLevel(logrus.TraceLevel)
	}
}

// Register the daemon command.
func Register(app *kingpin.Application) {
	registerDaemon(app)
	registerProcess(app)
}

func registerDaemon(app *kingpin.Application) {
	c := new(daemonCommand)

	cmd := app.Command("daemon", "starts the runner daemon").
		Default().
		Action(c.run)

	cmd.Arg("envfile", "load the environment variable file").
		Default("").
		StringVar(&c.envfile)
}
