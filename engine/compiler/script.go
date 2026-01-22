// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package compiler

import (
	"github.com/drone-runners/drone-runner-docker/engine"
	"github.com/drone-runners/drone-runner-docker/engine/compiler/shell"
	"github.com/drone-runners/drone-runner-docker/engine/compiler/shell/powershell"
	"github.com/drone-runners/drone-runner-docker/engine/compiler/shell/wincmd"
	"github.com/drone-runners/drone-runner-docker/engine/resource"
)

// helper function configures the pipeline script for the
// target operating system.
func setupScript(src *resource.Step, dst *engine.Step, os string) {
	if len(src.Commands) > 0 {
		switch os {
		case "windows":
			setupScriptWindowsCmd(src, dst)
		default:
			setupScriptPosix(src, dst)
		}
	}
}

// helper function configures the pipeline script for the
// windows operating system.
// WIP, BROKEN
func setupScriptWindowsCmd(src *resource.Step, dst *engine.Step) {
	dst.Entrypoint = []string{"cmd", "/S", "/c"}
	dst.Command = []string{`\drone\drone-build-script.cmd`}
	dst.Envs["DRONE_SCRIPT"] = wincmd.Script(src.Commands)
	dst.Envs["SHELL"] = "cmd.exe"
}

// helper function configures the pipeline script for the
// windows operating system.
func setupScriptWindowsPowershell(src *resource.Step, dst *engine.Step) {
	dst.Entrypoint = []string{"powershell", "-noprofile", "-noninteractive", "-command"}
	dst.Command = []string{"echo $Env:DRONE_SCRIPT | iex"}
	dst.Envs["DRONE_SCRIPT"] = powershell.Script(src.Commands)
	dst.Envs["SHELL"] = "powershell.exe"
}

// helper function configures the pipeline script for the
// windows operating system.
func setupScriptWindowsPwsh(src *resource.Step, dst *engine.Step) {
	dst.Entrypoint = []string{"pwsh", "-noprofile", "-noninteractive", "-command"}
	dst.Command = []string{"echo $Env:DRONE_SCRIPT | iex"}
	dst.Envs["DRONE_SCRIPT"] = powershell.Script(src.Commands)
	dst.Envs["SHELL"] = "pwsh.exe"
}

// helper function configures the pipeline script for the
// linux operating system.
func setupScriptPosix(src *resource.Step, dst *engine.Step) {
	dst.Entrypoint = []string{"/bin/sh", "-c"}
	dst.Command = []string{`echo "$DRONE_SCRIPT" | /bin/sh`}
	dst.Envs["DRONE_SCRIPT"] = shell.Script(src.Commands)
}
