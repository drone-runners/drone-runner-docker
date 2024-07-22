// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package compiler

import (
	stdpath "path"
	"strings"

	"github.com/drone-runners/drone-runner-docker/engine"
	"github.com/drone-runners/drone-runner-docker/engine/resource"
)

const (
	workspacePath     = "/drone/src"
	workspaceName     = "workspace"
	workspaceHostName = "host"
)

func createWorkspace(from *resource.Pipeline) (base, path, full string) {
	base = from.Workspace.Base
	path = from.Workspace.Path
	if base == "" {
		if strings.HasPrefix(path, "/") {
			base = path
			path = ""
		} else {
			base = workspacePath
		}
	}
	full = stdpath.Join(base, path)

	if from.Platform.OS == "windows" {
		base = toWindowsDrive(base)
		path = toWindowsPath(path)
		full = toWindowsDrive(full)
	}
	return base, path, full
}

func setupWorkdir(src *resource.Step, dst *engine.Step, path string) {
	// if the working directory is already set
	// do not alter.
	if dst.WorkingDir != "" {
		return
	}
	// if the user is running the container as a
	// service (detached mode) with no commands, we
	// should use the default working directory.
	if dst.Detach && len(src.Commands) == 0 {
		return
	}
	// else set the working directory.
	dst.WorkingDir = path
}

// helper function appends the workspace base and
// path to the step's list of environment variables.
func setupWorkspaceEnv(step *engine.Step, base, path, full string) {
	step.Envs["DRONE_WORKSPACE_BASE"] = base
	step.Envs["DRONE_WORKSPACE_PATH"] = path
	step.Envs["DRONE_WORKSPACE"] = full
	step.Envs["CI_WORKSPACE_BASE"] = base
	step.Envs["CI_WORKSPACE_PATH"] = path
	step.Envs["CI_WORKSPACE"] = full
}

// helper function converts the path to a valid windows
// path, including the default C drive.
func toWindowsDrive(s string) string {
	return "c:" + toWindowsPath(s)
}

// helper function converts the path to a valid windows
// path, replacing backslashes with forward slashes.
func toWindowsPath(s string) string {
	return strings.Replace(s, "/", "\\", -1)
}
