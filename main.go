// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"github.com/earthly/earthly/buildkitd"
	"github.com/earthly/earthly/conslogging"
	"github.com/earthly/earthly/domain"
	"github.com/earthly/earthly/util/llbutil"
	"github.com/earthly/earthly/util/llbutil/pllb"
	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/client/llb"
	"golang.org/x/sync/errgroup"
	"io"

	// "github.com/earthly/earthly/util/llbutil/pllb"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	//command.Command()
	//targets, _ := earthfile2llb.GetTargets("Earthfile")
	target, _ := domain.ParseTarget("+build")
	fmt.Print(target)
	//fmt.Print(targets)

	console := conslogging.Current(conslogging.ForceColor, conslogging.DefaultPadding, false)
	// Bootstrap buildkit - pulls image and starts daemon.
	//ctx, _ := context.WithTimeout(ctx, 100*time.Millisecond)
	buildkitdImage := "earthly/buildkitd:main"
	ctx := context.Background()
	bkClient, _ := buildkitd.NewClient(ctx, console, buildkitdImage, buildkitd.Settings{BuildkitAddress: "docker-container://earthly-buildkitd", DebuggerAddress: "tcp://127.0.0.1:8373", LocalRegistryAddress: "tcp://127.0.0.1:8371", UseTCP: false, UseTLS: false})
	state := pllb.Image("earthly/buildkitd:main", llb.Platform(llbutil.DefaultPlatform()))

	dt, _ := state.Marshal(ctx, llb.Platform(llbutil.DefaultPlatform()))
	solveOpt := &client.SolveOpt{
		Exports: []client.ExportEntry{
			{
				Type: client.ExporterDocker,
				Attrs: map[string]string{
					"name":                  "",
					"containerimage.config": "",
				},
				Output: func(_ map[string]string) (io.WriteCloser, error) {
					return nil, nil
				},
			},
		},
	}
	ch := make(chan *client.SolveStatus)
	con := conslogging.Current(conslogging.ForceColor, conslogging.DefaultPadding, false)
	sm := newSolverMonitor(con, true, true)
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		_, _ = bkClient.Solve(ctx, dt, *solveOpt, ch)
		return nil
	})
	sm.PrintTiming()
	var vertexFailureOutput string

	eg.Go(func() error {
		var err error
		vertexFailureOutput, err = sm.monitorProgress(ctx, ch, "", true)
		return err
	})
	eg.Wait()
	fmt.Print(vertexFailureOutput)
	fmt.Print(bkClient)
}
