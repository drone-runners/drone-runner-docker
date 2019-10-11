// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package compiler

import (
	"reflect"
	"testing"

	"github.com/drone/runner-go/shell/bash"
	"github.com/drone/runner-go/shell/powershell"

	"github.com/dchest/uniuri"
)

func Test_tempdir(t *testing.T) {
	// replace the default random function with one that
	// is deterministic, for testing purposes.
	random = notRandom

	// restore the default random function and the previously
	// specified temporary directory
	defer func() {
		random = uniuri.New
	}()

	tests := []struct {
		os   string
		path string
	}{
		{os: "windows", path: "C:\\Windows\\Temp\\drone-random"},
		{os: "linux", path: "/tmp/drone-random"},
		{os: "openbsd", path: "/tmp/drone-random"},
		{os: "netbsd", path: "/tmp/drone-random"},
		{os: "freebsd", path: "/tmp/drone-random"},
	}

	for _, test := range tests {
		if got, want := tempdir(test.os), test.path; got != want {
			t.Errorf("Want tempdir %s, got %s", want, got)
		}
	}
}

func Test_join(t *testing.T) {
	tests := []struct {
		os string
		a  []string
		b  string
	}{
		{os: "windows", a: []string{"C:", "Windows", "Temp"}, b: "C:\\Windows\\Temp"},
		{os: "linux", a: []string{"/tmp", "foo", "bar"}, b: "/tmp/foo/bar"},
	}
	for _, test := range tests {
		if got, want := join(test.os, test.a...), test.b; got != want {
			t.Errorf("Want %s, got %s", want, got)
		}
	}
}

func Test_getExt(t *testing.T) {
	tests := []struct {
		os string
		a  string
		b  string
	}{
		{os: "windows", a: "clone", b: "clone.ps1"},
		{os: "linux", a: "clone", b: "clone"},
	}
	for _, test := range tests {
		if got, want := getExt(test.os, test.a), test.b; got != want {
			t.Errorf("Want %s, got %s", want, got)
		}
	}
}

func Test_getCommand(t *testing.T) {
	cmd, args := getCommand("linux", "clone.sh")
	if got, want := cmd, "/bin/sh"; got != want {
		t.Errorf("Want command %s, got %s", want, got)
	}
	if !reflect.DeepEqual(args, []string{"-e", "clone.sh"}) {
		t.Errorf("Unexpected args %v", args)
	}

	cmd, args = getCommand("windows", "clone.ps1")
	if got, want := cmd, "powershell"; got != want {
		t.Errorf("Want command %s, got %s", want, got)
	}
	if !reflect.DeepEqual(args, []string{"-noprofile", "-noninteractive", "-command", "clone.ps1"}) {
		t.Errorf("Unexpected args %v", args)
	}
}

func Test_getNetrc(t *testing.T) {
	tests := []struct {
		os   string
		name string
	}{
		{os: "windows", name: "_netrc"},
		{os: "linux", name: ".netrc"},
		{os: "openbsd", name: ".netrc"},
		{os: "netbsd", name: ".netrc"},
		{os: "freebsd", name: ".netrc"},
	}
	for _, test := range tests {
		if got, want := getNetrc(test.os), test.name; got != want {
			t.Errorf("Want %s on %s, got %s", want, test.os, got)
		}
	}
}

func Test_getScript(t *testing.T) {
	commands := []string{"go build"}

	a := genScript("windows", commands)
	b := powershell.Script(commands)
	if !reflect.DeepEqual(a, b) {
		t.Errorf("Generated windows linux script")
	}

	a = genScript("linux", commands)
	b = bash.Script(commands)
	if !reflect.DeepEqual(a, b) {
		t.Errorf("Generated invalid linux script")
	}
}
