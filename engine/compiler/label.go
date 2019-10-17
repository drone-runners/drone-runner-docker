// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package compiler

import (
	"fmt"
	"time"

	"github.com/drone/drone-go/drone"
)

func createLabels(
	repo *drone.Repo,
	build *drone.Build,
	stage *drone.Stage,
	step *drone.Step,
) map[string]string {
	labels := map[string]string{
		"io.drone":                "true",
		"io.drone.build.number":   fmt.Sprint(build.Number),
		"io.drone.repo.namespace": repo.Namespace,
		"io.drone.repo.name":      repo.Name,
		"io.drone.stage.name":     stage.Name,
		"io.drone.stage.number":   fmt.Sprint(stage.Number),
		"io.drone.ttl":            fmt.Sprint(time.Duration(repo.Timeout) * time.Minute),
		"io.drone.expires":        fmt.Sprint(time.Now().Add(time.Duration(repo.Timeout)*time.Minute + time.Hour).Unix()),
		"io.drone.created":        fmt.Sprint(time.Now().Unix()),
		"io.drone.protected":      "false",
	}
	if step != nil {
		labels["io.drone.step.name"] = step.Name
		labels["io.drone.step.number"] = fmt.Sprint(step.Number)
	}
	return labels
}
