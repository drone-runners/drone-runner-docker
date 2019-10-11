// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package engine

import (
	"context"
	"io"
)

// New returns a new engine.
func New(publickeyFile, privatekeyFile string) (Engine, error) {
	return &engine{}, nil
}

type engine struct {
}

// Setup the pipeline environment.
func (e *engine) Setup(ctx context.Context, spec *Spec) error {
	return nil
}

// Destroy the pipeline environment.
func (e *engine) Destroy(ctx context.Context, spec *Spec) error {
	return nil
}

// Run runs the pipeline step.
func (e *engine) Run(ctx context.Context, spec *Spec, step *Step, output io.Writer) (*State, error) {
	return nil, nil
}
