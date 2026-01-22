// Package wincmd provides functions for converting shell
// commands to cmd scripts.
package wincmd

import (
	"bytes"
)

// Script converts a slice of individual shell commands to
// a cmd script.
func Script(commands []string) string {
	buf := new(bytes.Buffer)

	// ignore first linebreak
	buf.WriteString(optionScript)

	for _, command := range commands {
		buf.WriteString("echo +--------------------------------\n")
		buf.WriteString("echo + " + command + "\n")
		buf.WriteString(command)
		buf.WriteString(`
set LastErrorLevel=%ErrorLevel%
if %LastErrorLevel% gtr 0 (
  echo ERROR: %LastErrorLevel% 
  exit /b %LastErrorLevel%
)` + "\n")
		buf.WriteString("echo.\n")
	}

	return buf.String()
}

// optionScript is a helper script this is added to the build
// to set shell options, in this case, to exit on error.
const optionScript = `
@echo off
if .%DRONE_NETRC_MACHINE%. neq .. (
  type nul > %USERPROFILE%\.netrc
  echo machine %DRONE_NETRC_MACHINE% >> %USERPROFILE%\.netrc
  echo login %DRONE_NETRC_USERNAME% >> %USERPROFILE%\.netrc
  echo password %DRONE_NETRC_PASSWORD% >> %USERPROFILE%\.netrc
)
set DRONE_NETRC_USERNAME=
set DRONE_NETRC_PASSWORD=
set DRONE_NETRC_FILE=
set DRONE_SCRIPT=

`
