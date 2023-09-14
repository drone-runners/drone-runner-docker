// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package engine

import (
	"context"
	"io"
)

type (
	// Engine is the interface that must be implemented by a
	// pipeline execution engine.
	Engine interface {
		// Setup the pipeline environment.
		Setup(context.Context, *Spec) error

		// Destroy the pipeline environment.
		Destroy(context.Context, *Spec) error

		// Run runs the pipeline step.
		Run(context.Context, *Spec, *Step, io.Writer) (*State, error)
	}

	// State reports the execution state.
	State struct {
		// ExitCode returns the exit code of the exited step.
		ExitCode int

		// GetExited reports whether the step has exited.
		Exited bool

		// OOMKilled reports whether the step has been
		// killed by the process manager.
		OOMKilled bool
	}
)
