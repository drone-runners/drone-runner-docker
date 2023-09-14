// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package compiler

import (
	"github.com/drone-runners/drone-runner-docker/engine/compiler/shell"
	"github.com/drone-runners/drone-runner-docker/engine/compiler/shell/powershell"
	"github.com/drone-runners/drone-runner-docker/engine/experimental/engine"
)

// helper function configures the pipeline script for the
// target operating system.
func setupScript(dst *engine.Step, script, os string) {
	if script != "" {
		switch os {
		case "windows":
			setupScriptWindows(dst, script)
		default:
			setupScriptPosix(dst, script)
		}
	}
}

// helper function configures the pipeline script for the
// windows operating system.
func setupScriptWindows(dst *engine.Step, commands ...string) {
	dst.Entrypoint = []string{"powershell", "-noprofile", "-noninteractive", "-command"}
	dst.Command = []string{"echo $Env:DRONE_SCRIPT | iex"}
	dst.Envs["DRONE_SCRIPT"] = powershell.Script(commands)
	dst.Envs["SHELL"] = "powershell.exe"
}

// helper function configures the pipeline script for the
// linux operating system.
func setupScriptPosix(dst *engine.Step, commands ...string) {
	dst.Entrypoint = []string{"/bin/sh", "-c"}
	dst.Command = []string{`echo "$DRONE_SCRIPT" | /bin/sh`}
	dst.Envs["DRONE_SCRIPT"] = shell.Script(commands)
}
