// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package delegate

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/drone-runners/drone-runner-docker/engine/resource"

	"github.com/drone-runners/drone-runner-docker/engine"
	"github.com/drone-runners/drone-runner-docker/command/delegate/livelog"
	loghistory "github.com/drone/runner-go/logger/history"
	"github.com/drone/runner-go/server"
	"github.com/drone/signal"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"gopkg.in/alecthomas/kingpin.v2"
)

type delegateCommand struct {
	envfile string
}

func (c *delegateCommand) run(*kingpin.ParseContext) error {
	// load environment variables from file.
	//godotenv.Load(c.envfile)

	logrus.SetLevel(logrus.TraceLevel)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// listen for termination signals to gracefully shutdown
	// the runner daemon.
	ctx = signal.WithContextFunc(ctx, func() {
		println("received signal, terminating process")
		cancel()
	})

	//cli := client.New(
	//	config.Client.Address,
	//	config.Client.Secret,
	//	config.Client.SkipVerify,
	//)
	//if config.Client.Dump {
	//	cli.Dumper = logger.StandardDumper(
	//		config.Client.DumpBody,
	//	)
	//}
	//cli.Logger = logger.Logrus(
	//	logrus.NewEntry(
	//		logrus.StandardLogger(),
	//	),
	//)

	opts := engine.Opts{
		//HidePull: !config.Docker.Stream,
		HidePull: false,
	}
	engineInstance, err := engine.NewEnv(opts)
	if err != nil {
		logrus.WithError(err).
			Fatalln("cannot load the docker engine")
	}
	for {
		err := engineInstance.Ping(ctx)
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

	//remote := remote.New(cli)
	//tracer := history.New(remote)
	hook := loghistory.New()
	logrus.AddHook(hook)

	// runner := &runtime.Runner{
	// 	Client:   cli,
	// 	Machine:  config.Runner.Name,
	// 	Environ:  config.Runner.Environ,
	// 	Reporter: tracer,
	// 	Lookup:   resource.Lookup,
	// 	Lint:     linter.New().Lint,
	// 	Match: match.Func(
	// 		config.Limit.Repos,
	// 		config.Limit.Events,
	// 		config.Limit.Trusted,
	// 	),
	// 	Compiler: &compiler.Compiler{
	// 		Clone:          config.Runner.Clone,
	// 		Privileged:     append(config.Runner.Privileged, compiler.Privileged...),
	// 		Networks:       config.Runner.Networks,
	// 		NetworkOpts:    config.Runner.NetworkOpts,
	// 		NetrcCloneOnly: config.Netrc.CloneOnly,
	// 		Volumes:        config.Runner.Volumes,
	// 		Resources: compiler.Resources{
	// 			Memory:     config.Resources.Memory,
	// 			MemorySwap: config.Resources.MemorySwap,
	// 			CPUQuota:   config.Resources.CPUQuota,
	// 			CPUPeriod:  config.Resources.CPUPeriod,
	// 			CPUShares:  config.Resources.CPUShares,
	// 			CPUSet:     config.Resources.CPUSet,
	// 			ShmSize:    config.Resources.ShmSize,
	// 		},
	// 		Tmate: compiler.Tmate{
	// 			Image:          config.Tmate.Image,
	// 			Enabled:        config.Tmate.Enabled,
	// 			Server:         config.Tmate.Server,
	// 			Port:           config.Tmate.Port,
	// 			RSA:            config.Tmate.RSA,
	// 			ED25519:        config.Tmate.ED25519,
	// 			AuthorizedKeys: config.Tmate.AuthorizedKeys,
	// 		},
	// 		Environ: provider.Combine(
	// 			provider.Static(config.Runner.Environ),
	// 			provider.External(
	// 				config.Environ.Endpoint,
	// 				config.Environ.Token,
	// 				config.Environ.SkipVerify,
	// 			),
	// 		),
	// 		Registry: registry.Combine(
	// 			registry.File(
	// 				config.Docker.Config,
	// 			),
	// 			registry.External(
	// 				config.Registry.Endpoint,
	// 				config.Registry.Token,
	// 				config.Registry.SkipVerify,
	// 			),
	// 		),
	// 		Secret: secret.Combine(
	// 			secret.StaticVars(
	// 				config.Runner.Secrets,
	// 			),
	// 			secret.External(
	// 				config.Secret.Endpoint,
	// 				config.Secret.Token,
	// 				config.Secret.SkipVerify,
	// 			),
	// 		),
	// 	},
	// 	Exec: runtime.NewExecer(
	// 		tracer,
	// 		remote,
	// 		engineInstance,
	// 		config.Runner.Procs,
	// 	).Exec,
	// }

	// poller := &poller.Poller{
	// 	Client:   cli,
	// 	Dispatch: runner.Run,
	// 	Filter: &client.Filter{
	// 		Kind:    resource.Kind,
	// 		Type:    resource.Type,
	// 		OS:      config.Platform.OS,
	// 		Arch:    config.Platform.Arch,
	// 		Variant: config.Platform.Variant,
	// 		Kernel:  config.Platform.Kernel,
	// 		Labels:  config.Runner.Labels,
	// 	},
	// }

	var g errgroup.Group
	server := server.Server{
		Addr:    ":3000", // config.Server.Port,
		Handler: delegateListener(engineInstance),
	}

	logrus.WithField("addr", ":3000" /*config.Server.Port*/).
		Infoln("starting the server")

	g.Go(func() error {
		return server.ListenAndServe(ctx)
	})

	// Ping the server and block until a successful connection
	// to the server has been established.
	// for {
	// 	err := cli.Ping(ctx, config.Runner.Name)
	// 	select {
	// 	case <-ctx.Done():
	// 		return nil
	// 	default:
	// 	}
	// 	if ctx.Err() != nil {
	// 		break
	// 	}
	// 	if err != nil {
	// 		logrus.WithError(err).
	// 			Errorln("cannot ping the remote server")
	// 		time.Sleep(time.Second)
	// 	} else {
	// 		logrus.Infoln("successfully pinged the remote server")
	// 		break
	// 	}
	// }

	g.Go(func() error {
		logrus.WithField("capacity", 2 /*config.Runner.Capacity*/).
			//WithField("endpoint", config.Client.Address).
			WithField("kind", resource.Kind).
			WithField("type", resource.Type).
			WithField("os", "linux" /*config.Platform.OS*/).
			WithField("arch", "amd64" /*config.Platform.Arch*/).
			Infoln("polling the remote server")

		//poller.Poll(ctx, config.Runner.Capacity)
		return nil
	})

	err = g.Wait()
	if err != nil {
		logrus.WithError(err).
			Errorln("shutting down the server")
	}
	return err
}

func delegateListener(engine *engine.Docker) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/setup", handleSetup(engine))
	mux.HandleFunc("/destroy", handleDestroy(engine))
	mux.HandleFunc("/step", handleStep(engine))

	// // omit dashboard handlers when no password configured.
	// if config.Password == "" {
	// 	return mux
	// }

	// // middleware to require basic authentication.
	// auth := basicauth.New(config.Realm, map[string][]string{
	// 	config.Username: {config.Password},
	// })

	// // handler to serve static assets for the dashboard.
	// fs := http.FileServer(static.New())

	// // dashboard handles.
	// mux.Handle("/static/", http.StripPrefix("/static/", fs))
	// mux.Handle("/logs", auth(handler.HandleLogHistory(history)))
	// mux.Handle("/view", auth(handler.HandleStage(tracer, history)))
	// mux.Handle("/", auth(handler.HandleIndex(tracer)))
	return mux
}

func handleSetup(eng *engine.Docker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		reqData, err := GetSetupRequest(r.Body)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		stageID := reqData.StageID

		spec, err := CompileDelegateStage()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = Stages.Store(stageID, spec)
		if err != nil {
			logrus.WithError(err).
				Errorln("failed to store spec")
			w.WriteHeader(http.StatusInternalServerError)
		}

		err = eng.Setup(r.Context(), spec)
		if err != nil {
			logrus.WithError(err).
				Errorln("cannot setup the docker environment")
			w.WriteHeader(http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusOK)
	}
}

func handleStep(eng *engine.Docker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		reqData, err := GetExecStepRequest(r.Body)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		stageID := reqData.StageID

		spec, err := Stages.Get(stageID)
		if err != nil {
			logrus.WithError(err).
				Errorln("failed to get the stage")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if spec == nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		stepID := random()

		// create a step to run, why do we do this ? why not use the engine.spec
		steppy := engine.Step{
			ID:         stepID,
			Name:       reqData.StepID,
			WorkingDir: "/drone/src",
			Command:    []string{reqData.Command},
			Entrypoint: []string{"/bin/sh", "-c"},
			Image:      reqData.Image,
		}

		c := livelog.NewHTTPClient("http://localhost:8079", "accountID", "token", true)

		// create a writer
		wc := livelog.New(c, reqData.Dump.Step.LogKey)
		defer wc.Close()
		state, err := eng.Run(r.Context(), spec, &steppy, wc)
		if err != nil {
			logrus.WithError(err).
				Errorln("running the step failed. this is a runner error")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_ = json.NewEncoder(w).Encode(state)
	}
}

func handleDestroy(eng *engine.Docker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		reqData, err := GetDestroyRequest(r.Body)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		stageID := reqData.StageID

		spec, err := Stages.Get(stageID)
		if err != nil {
			logrus.WithError(err).
				Errorln("failed to delete the stage")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = eng.Destroy(r.Context(), spec)
		if err != nil {
			logrus.WithError(err).
				Errorln("cannot destroy the docker environment")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, _ = Stages.Remove(stageID)

		w.WriteHeader(http.StatusOK)
	}
}
func RegisterDelegate(app *kingpin.Application) {
	c := new(delegateCommand)

	cmd := app.Command("delegate", "starts the delegate").
		Action(c.run)

	cmd.Arg("envfile", "load the environment variable file").
		Default("").
		StringVar(&c.envfile)
}
