// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package compiler

import (
	"fmt"
	"strings"

	"github.com/drone/runner-go/shell/bash"
	"github.com/drone/runner-go/shell/powershell"
)

// helper function returns the base temporary directory based
// on the target platform.
func tempdir(os string) string {
	dir := fmt.Sprintf("drone-%s", random())
	switch os {
	case "windows":
		return join(os, "C:\\Windows\\Temp", dir)
	default:
		return join(os, "/tmp", dir)
	}
}

// helper function joins the file paths.
func join(os string, paths ...string) string {
	switch os {
	case "windows":
		return strings.Join(paths, "\\")
	default:
		return strings.Join(paths, "/")
	}
}

// helper function returns the shell extension based on the
// target platform.
func getExt(os, file string) (s string) {
	switch os {
	case "windows":
		return file + ".ps1"
	default:
		return file
	}
}

//
// TODO(bradrydzewski) can we remove the below functions?
//

// helper function returns the shell command and arguments
// based on the target platform to invoke the script
func getCommand(os, script string) (string, []string) {
	cmd, args := bash.Command()
	switch os {
	case "windows":
		cmd, args = powershell.Command()
	}
	return cmd, append(args, script)
}

// helper function returns the netrc file name based on the
// target platform.
func getNetrc(os string) string {
	switch os {
	case "windows":
		return "_netrc"
	default:
		return ".netrc"
	}
}

// helper function generates and returns a shell script to
// execute the provided shell commands. The shell scripting
// language (bash vs pwoershell) is determined by the operating
// system.
func genScript(os string, commands []string) string {
	switch os {
	case "windows":
		return powershell.Script(commands)
	default:
		return bash.Script(commands)
	}
}
