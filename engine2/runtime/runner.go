// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package runtime

import (
	"context"
	"regexp"
	"time"

	"github.com/drone-runners/drone-runner-docker/engine2/compiler"
	"github.com/drone-runners/drone-runner-docker/engine2/engine"
	"github.com/drone-runners/drone-runner-docker/engine2/inputs"

	legacy "github.com/drone/runner-go/pipeline/runtime"

	"github.com/drone/runner-go/client"
	"github.com/drone/runner-go/logger"
	"github.com/drone/runner-go/pipeline"
	"github.com/drone/runner-go/secret"

	"github.com/drone/drone-go/drone"

	harness "github.com/drone/spec/dist/go"
	"github.com/drone/spec/dist/go/parse/expand"
	"github.com/drone/spec/dist/go/parse/normalize"
	"github.com/drone/spec/dist/go/parse/script"
	"github.com/drone/spec/dist/go/parse/walk"
)

var noContext = context.Background()

// Runner runs the pipeline.
type Runner struct {
	// Machine provides the runner with the name of the host
	// machine executing the pipeline.
	Machine string

	// Client is the remote client responsible for interacting
	// with the central server.
	Client client.Client

	// Compiler is responsible for compiling the pipeline
	// configuration to the intermediate representation.
	Compiler compiler.Compiler

	// Reporter reports pipeline status and logs back to the
	// remote server.
	Reporter pipeline.Reporter

	// Execer is responsible for executing intermediate
	// representation of the pipeline and returns its results.
	Exec func(context.Context, *engine.Spec, *pipeline.State) error

	// LegacyRunner is used to run a legacy yaml configuration
	// file from drone.
	LegacyRunner *legacy.Runner
}

// Run runs the pipeline stage.
func (s *Runner) Run(ctx context.Context, stage *drone.Stage) error {
	log := logger.FromContext(ctx).
		WithField("stage.id", stage.ID).
		WithField("stage.name", stage.Name).
		WithField("stage.number", stage.Number)

	log.Debug("stage received")

	// delivery to a single agent is not guaranteed, which means
	// we need confirm receipt. The first agent that confirms
	// receipt of the stage can assume ownership.

	stage.Machine = s.Machine
	err := s.Client.Accept(ctx, stage)
	if err != nil && err == client.ErrOptimisticLock {
		log.Debug("stage accepted by another runner")
		return nil
	}
	if err != nil {
		log.WithError(err).Error("cannot accept stage")
		return err
	}

	log.Debug("stage accepted")

	data, err := s.Client.Detail(ctx, stage)
	if err != nil {
		log.WithError(err).Error("cannot get stage details")
		return err
	}

	log = log.WithField("repo.id", data.Repo.ID).
		WithField("repo.namespace", data.Repo.Namespace).
		WithField("repo.name", data.Repo.Name).
		WithField("build.id", data.Build.ID).
		WithField("build.number", data.Build.Number)

	log.Debug("stage details fetched")

	// if we are dealing with the legacy drone yaml, use
	// the legacy drone engine.
	if !regexp.MustCompilePOSIX(`^spec:`).Match(data.Config.Data) {
		return s.LegacyRunner.RunAccepted(ctx, stage.ID)
	}

	return s.run(ctx, stage, data)
}

func (s *Runner) run(ctx context.Context, stage *drone.Stage, data *client.Context) error {
	log := logger.FromContext(ctx).
		WithField("repo.id", data.Repo.ID).
		WithField("stage.id", stage.ID).
		WithField("stage.name", stage.Name).
		WithField("stage.number", stage.Number).
		WithField("repo.namespace", data.Repo.Namespace).
		WithField("repo.name", data.Repo.Name).
		WithField("build.id", data.Build.ID).
		WithField("build.number", data.Build.Number)

	ctxdone, cancel := context.WithCancel(ctx)
	defer cancel()

	timeout := time.Duration(data.Repo.Timeout) * time.Minute
	ctxtimeout, cancel := context.WithTimeout(ctxdone, timeout)
	defer cancel()

	ctxcancel, cancel := context.WithCancel(ctxtimeout)
	defer cancel()

	// next we opens a connection to the server to watch for
	// cancellation requests. If a build is cancelled the running
	// stage should also be cancelled.
	go func() {
		done, _ := s.Client.Watch(ctxdone, data.Build.ID)
		if done {
			cancel()
			log.Debugln("received cancellation")
		} else {
			log.Debugln("done listening for cancellations")
		}
	}()

	// envs := environ.Combine(
	// 	s.Environ,
	// 	environ.System(data.System),
	// 	environ.Repo(data.Repo),
	// 	environ.Build(data.Build),
	// 	environ.Stage(stage),
	// 	environ.Link(data.Repo, data.Build, data.System),
	// 	data.Build.Params,
	// )

	// // string substitution function ensures that string
	// // replacement variables are escaped and quoted if they
	// // contain a newline character.
	// subf := func(k string) string {
	// 	v := envs[k]
	// 	if strings.Contains(v, "\n") {
	// 		v = fmt.Sprintf("%q", v)
	// 	}
	// 	return v
	// }

	state := &pipeline.State{
		Build:  data.Build,
		Stage:  stage,
		Repo:   data.Repo,
		System: data.System,
	}

	config, err := harness.ParseBytes(data.Config.Data)
	if err != nil {
		log.WithError(err).Error("cannot parse configuration file")
		state.FailAll(err)
		return s.Reporter.ReportStage(noContext, state)
	}

	// expand matrix stages and steps
	expand.Expand(config)

	// expand expressions for the stage and step
	// name and identifier fields only.
	inputParams := map[string]interface{}{}
	inputParams["repo"] = inputs.Repo(data.Repo)
	inputParams["build"] = inputs.Build(data.Build)
	// TODO add inputsParams["stage"]
	walk.Walk(config, func(v interface{}) error {
		switch vv := v.(type) {
		case *harness.Pipeline:
			inputParams["inputs"] = inputs.Inputs(vv.Inputs, data.Build.Params)
		case *harness.Step:
			if vv.Strategy != nil && vv.Strategy.Spec != nil {
				if matrix, ok := vv.Strategy.Spec.(*harness.Matrix); ok {
					for _, axis := range matrix.Include {
						inputParams["matrix"] = axis
					}
				}
			}
			vv.Id = script.Expand(vv.Id, inputParams)
			vv.Name = script.Expand(vv.Name, inputParams)
		case *harness.Stage:
			if vv.Strategy != nil && vv.Strategy.Spec != nil {
				if matrix, ok := vv.Strategy.Spec.(*harness.Matrix); ok {
					for _, axis := range matrix.Include {
						inputParams["matrix"] = axis
					}
				}
			}
			vv.Id = script.Expand(vv.Id, inputParams)
			vv.Name = script.Expand(vv.Name, inputParams)
		}
		return nil
	})

	// normalize the configuration to ensure
	// all steps have an identifier
	normalize.Normalize(config)

	// compile the yaml configuration file to an intermediate
	// representation, and then
	args := compiler.Args{
		Config: config,
		Build:  data.Build,
		Stage:  stage,
		Repo:   data.Repo,
		System: data.System,
		Netrc:  data.Netrc,
		Secret: secret.Static(data.Secrets),
	}

	spec, err := s.Compiler.Compile(ctx, args)
	if err != nil {
		log.WithError(err).Error("cannot find pipeline or stage resource in yaml")
		state.FailAll(err)
		return s.Reporter.ReportStage(noContext, state)
	}

	for _, src := range spec.Steps {

		// steps that are skipped are ignored and are not stored
		// in the drone database, nor displayed in the UI.
		if src.RunPolicy == engine.RunNever {
			continue
		}
		stage.Steps = append(stage.Steps, &drone.Step{
			Name:      src.Name,
			Number:    len(stage.Steps) + 1,
			StageID:   stage.ID,
			Status:    drone.StatusPending,
			ErrIgnore: src.ErrPolicy == engine.ErrIgnore,
			Image:     src.Image,
			Detached:  src.Detach,
			DependsOn: src.DependsOn,
		})
	}

	stage.Started = time.Now().Unix()
	stage.Status = drone.StatusRunning
	if err := s.Client.Update(ctx, stage); err != nil {
		log.WithError(err).Error("cannot update stage")
		return err
	}

	log.Debug("updated stage to running")

	ctxlogger := logger.WithContext(ctxcancel, log)
	if err := s.Exec(ctxlogger, spec, state); err != nil {
		log.WithError(err).
			WithField("duration", stage.Stopped-stage.Started).
			Debug("stage failed")
		return err
	}
	log.WithField("duration", stage.Stopped-stage.Started).
		Debug("updated stage to complete")
	return nil
}
