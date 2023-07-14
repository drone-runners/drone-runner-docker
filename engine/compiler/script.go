// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package compiler

import (
	"github.com/drone-runners/drone-runner-docker/engine"
	"github.com/drone-runners/drone-runner-docker/engine/compiler/shell"
	"github.com/drone-runners/drone-runner-docker/engine/compiler/shell/powershell"
	"github.com/drone-runners/drone-runner-docker/engine/resource"
)

// helper function configures the pipeline script for the
// target operating system.
func setupScript(src *resource.Step, dst *engine.Step, os string) {
	if len(src.Commands) > 0 {
		switch os {
		case "windows":
			setupScriptWindows(src, dst)
		default:
			setupScriptPosix(src, dst)
		}
	}
}

// helper function configures the pipeline script for the
// windows operating system.
// WIP, BROKEN
func setupScriptWindowsCmd(src *resource.Step, dst *engine.Step) {

	//dst.Command = []string{
	//	"setlocal",
	//	"EnableDelayedExpansion", "&",
	//	"for", "/f", "%L", "in", `(%DRONE_SCRIPT%)`,
	//	"do", "(",
	//	"echo", "+", "%L",
	//	")", "&",
	//}

	//dst.Command = []string{"set/p", "_discard=%DRONE_SCRIPT%<nul>build_script.cmd"} // & call build_script.cmd
	//dst.Command = []string{`set/p_discard=%DRONE_SCRIPT%<nul>build_script.cmd & dir & echo ================== & type build_script.cmd & echo ================== `} // & call build_script.cmd
	//dst.Command = []string{"<nul", "(set/p_discard=%DRONE_SCRIPT%)>build_script.cmd"}

	//dst.Command = []string{"set"}
	//dst.Entrypoint = []string{"cmd", "/S", "/c", "<nul", "(set/p_discard=%DRONE_SCRIPT%)"}
	//dst.Entrypoint = []string{"cmd", "/S", "/c", "<nul", "(set/p_discard=%DRONE_SCRIPT%)>%DRONE_WORKSPACE%\\§§build§§.cmd"}
	//dst.Entrypoint = []string{"cmd", "/S", "/c", "echo", "%DRONE_WORKSPACE%\\§§build§§.cmd"}
	//dst.Entrypoint = []string{"cmd", "/S", "/c"}
	//dst.Command = []string{"<nul", "set", "/p", "_discard=%DRONE_SCRIPT%>%DRONE_WORKSPACE%__build__.cmd"}
	//dst.Command = []string{"<nul", "(set", "/p", "_discard=%DRONE_SCRIPT%)"}
	//dst.Command = []string{}
	dst.Envs["DRONE_SCRIPT"] = wincmd.Script(src.Commands)
	dst.Envs["SHELL"] = "cmd.exe"
}

	dst.Entrypoint = []string{"powershell", "-noprofile", "-noninteractive", "-command"}
	dst.Command = []string{"echo $Env:DRONE_SCRIPT | iex"}
	dst.Envs["DRONE_SCRIPT"] = powershell.Script(src.Commands)
	dst.Envs["SHELL"] = "powershell.exe"
}

// helper function configures the pipeline script for the
// linux operating system.
func setupScriptPosix(src *resource.Step, dst *engine.Step) {
	dst.Entrypoint = []string{"/bin/sh", "-c"}
	dst.Command = []string{`echo "$DRONE_SCRIPT" | /bin/sh`}
	dst.Envs["DRONE_SCRIPT"] = shell.Script(src.Commands)
}
