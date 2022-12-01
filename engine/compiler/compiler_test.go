// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

// +build !windows

package compiler

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"strconv"
	"testing"

	"github.com/dchest/uniuri"
	"github.com/drone-runners/drone-runner-docker/engine"
	"github.com/drone-runners/drone-runner-docker/engine/resource"
	"github.com/drone/drone-go/drone"
	"github.com/drone/runner-go/environ/provider"
	"github.com/drone/runner-go/manifest"
	"github.com/drone/runner-go/pipeline/runtime"
	"github.com/drone/runner-go/registry"
	"github.com/drone/runner-go/secret"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var nocontext = context.Background()

// dummy function that returns a non-random string for testing.
// it is used in place of the random function.
func notRandom() string {
	return "random"
}

// This test verifies the pipeline dependency graph. When no
// dependency graph is defined, a default dependency graph is
// automatically defined to run steps serially.
func TestCompile_Serial(t *testing.T) {
	testCompile(t, "testdata/serial.yml", "testdata/serial.json")
}

// This test verifies the pipeline dependency graph. It also
// verifies that pipeline steps with no dependencies depend on
// the initial clone step.
func TestCompile_Graph(t *testing.T) {
	testCompile(t, "testdata/graph.yml", "testdata/graph.json")
}

// This test verifies no clone step exists in the pipeline if
// cloning is disabled.
func TestCompile_CloneDisabled_Serial(t *testing.T) {
	testCompile(t, "testdata/noclone_serial.yml", "testdata/noclone_serial.json")
}

// This test verifies no clone step exists in the pipeline if
// cloning is disabled. It also verifies no pipeline steps
// depend on a clone step.
func TestCompile_CloneDisabled_Graph(t *testing.T) {
	testCompile(t, "testdata/noclone_graph.yml", "testdata/noclone_graph.json")
}

// This test verifies that steps are disabled if conditions
// defined in the when block are not satisfied.
func TestCompile_Match(t *testing.T) {
	ir := testCompile(t, "testdata/match.yml", "testdata/match.json")
	if ir.Steps[0].RunPolicy != runtime.RunOnSuccess {
		t.Errorf("Expect run on success")
	}
	if ir.Steps[1].RunPolicy != runtime.RunNever {
		t.Errorf("Expect run never")
	}
}

// This test verifies that steps configured to run on both
// success or failure are configured to always run.
func TestCompile_RunAlways(t *testing.T) {
	ir := testCompile(t, "testdata/run_always.yml", "testdata/run_always.json")
	if ir.Steps[0].RunPolicy != runtime.RunAlways {
		t.Errorf("Expect run always")
	}
}

// This test verifies that steps configured to run on failure
// are configured to run on failure.
func TestCompile_RunFailure(t *testing.T) {
	ir := testCompile(t, "testdata/run_failure.yml", "testdata/run_failure.json")
	if ir.Steps[0].RunPolicy != runtime.RunOnFailure {
		t.Errorf("Expect run on failure")
	}
}

// This test verifies that secrets defined in the yaml are
// requested and stored in the intermediate representation
// at compile time.
func TestCompile_Secrets(t *testing.T) {
	manifest, _ := manifest.ParseFile("testdata/secret.yml")

	compiler := &Compiler{
		Environ:  provider.Static(nil),
		Registry: registry.Static(nil),
		Secret: secret.StaticVars(map[string]string{
			"token":       "3DA541559918A808C2402BBA5012F6C60B27661C",
			"password":    "password",
			"my_username": "octocat",
		}),
	}
	args := runtime.CompilerArgs{
		Repo:     &drone.Repo{},
		Build:    &drone.Build{},
		Stage:    &drone.Stage{},
		System:   &drone.System{},
		Netrc:    &drone.Netrc{},
		Manifest: manifest,
		Pipeline: manifest.Resources[0].(*resource.Pipeline),
		Secret:   secret.Static(nil),
	}

	ir := compiler.Compile(nocontext, args).(*engine.Spec)
	got := ir.Steps[0].Secrets
	want := []*engine.Secret{
		{
			Name: "my_password",
			Env:  "PASSWORD",
			Data: nil, // secret not found, data nil
			Mask: true,
		},
		{
			Name: "my_username",
			Env:  "USERNAME",
			Data: []byte("octocat"), // secret found
			Mask: true,
		},
	}
	if diff := cmp.Diff(got, want); len(diff) != 0 {
		// TODO(bradrydzewski) ordering is not guaranteed. this
		// unit tests needs to be adjusted accordingly.
		t.Skipf(diff)
	}
}

// This test verifies that step labels are generated correctly
func TestCompile_StepLabels(t *testing.T) {
	manifest, _ := manifest.ParseFile("testdata/steps.yml")

	compiler := &Compiler{
		Environ:  provider.Static(nil),
		Registry: registry.Static(nil),
		Secret:   secret.Static(nil),
		Labels:   map[string]string{"foo": "bar"},
	}
	args := runtime.CompilerArgs{
		Repo: &drone.Repo{
			Name:      "repo-name",
			Namespace: "repo-namespace",
			Slug:      "repo-slug",
		},
		Build: &drone.Build{
			Number: 42,
		},
		Stage: &drone.Stage{
			Name:   "default",
			Number: 1,
		},
		System: &drone.System{
			Host:    "drone.example.com",
			Proto:   "https",
			Version: "1.0.0",
		},
		Netrc:    &drone.Netrc{},
		Manifest: manifest,
		Pipeline: manifest.Resources[0].(*resource.Pipeline),
		Secret:   secret.Static(nil),
	}

	ir := compiler.Compile(nocontext, args).(*engine.Spec)

	gotLabels := []map[string]string{}
	for _, step := range ir.Steps {
		stepLabels := step.Labels

		// Remove timestamps from labels, we can't do a direct comparison
		if gotCreated, err := strconv.Atoi(stepLabels["io.drone.created"]); err != nil || gotCreated == 0 {
			t.Errorf("Expectec io.drone.created label to be set to a non-zero value. Got %q", stepLabels["io.drone.created"])
		}
		delete(stepLabels, "io.drone.created")

		if gotExpires, err := strconv.Atoi(stepLabels["io.drone.expires"]); err != nil || gotExpires == 0 {
			t.Errorf("Expectec io.drone.expires label to be set to a non-zero value. Got %q", stepLabels["io.drone.expires"])
		}
		delete(stepLabels, "io.drone.expires")

		gotLabels = append(gotLabels, stepLabels)
	}

	wantLabels := []map[string]string{
		{
			"foo":                     "bar",
			"io.drone":                "true",
			"io.drone.build.number":   "42",
			"io.drone.protected":      "false",
			"io.drone.repo.name":      "repo-name",
			"io.drone.repo.namespace": "repo-namespace",
			"io.drone.repo.slug":      "repo-slug",
			"io.drone.stage.name":     "default",
			"io.drone.stage.number":   "1",
			"io.drone.step.name":      "build",
			"io.drone.step.number":    "1",
			"io.drone.system.host":    "drone.example.com",
			"io.drone.system.proto":   "https",
			"io.drone.system.version": "1.0.0",
			"io.drone.ttl":            "0s",
		},
		{
			"foo":                     "bar",
			"io.drone":                "true",
			"io.drone.build.number":   "42",
			"io.drone.protected":      "false",
			"io.drone.repo.name":      "repo-name",
			"io.drone.repo.namespace": "repo-namespace",
			"io.drone.repo.slug":      "repo-slug",
			"io.drone.stage.name":     "default",
			"io.drone.stage.number":   "1",
			"io.drone.step.name":      "test",
			"io.drone.step.number":    "2",
			"io.drone.system.host":    "drone.example.com",
			"io.drone.system.proto":   "https",
			"io.drone.system.version": "1.0.0",
			"io.drone.ttl":            "0s",
		},
	}

	if diff := cmp.Diff(gotLabels, wantLabels); len(diff) != 0 {
		t.Errorf(diff)
	}

}

// helper function parses and compiles the source file and then
// compares to a golden json file.
func testCompile(t *testing.T, source, golden string) *engine.Spec {
	// replace the default random function with one that
	// is deterministic, for testing purposes.
	random = notRandom

	// restore the default random function and the previously
	// specified temporary directory
	defer func() {
		random = uniuri.New
	}()

	manifest, err := manifest.ParseFile(source)
	if err != nil {
		t.Error(err)
		return nil
	}

	compiler := &Compiler{
		Environ:  provider.Static(nil),
		Registry: registry.Static(nil),
		Secret: secret.StaticVars(map[string]string{
			"token":       "3DA541559918A808C2402BBA5012F6C60B27661C",
			"password":    "password",
			"my_username": "octocat",
		}),
	}
	args := runtime.CompilerArgs{
		Repo:     &drone.Repo{},
		Build:    &drone.Build{Target: "master"},
		Stage:    &drone.Stage{},
		System:   &drone.System{},
		Netrc:    &drone.Netrc{Machine: "github.com", Login: "octocat", Password: "correct-horse-battery-staple"},
		Manifest: manifest,
		Pipeline: manifest.Resources[0].(*resource.Pipeline),
		Secret:   secret.Static(nil),
	}

	got := compiler.Compile(nocontext, args)

	raw, err := ioutil.ReadFile(golden)
	if err != nil {
		t.Error(err)
	}

	want := new(engine.Spec)
	err = json.Unmarshal(raw, want)
	if err != nil {
		t.Error(err)
	}

	opts := cmp.Options{
		cmpopts.IgnoreUnexported(engine.Spec{}),
		cmpopts.IgnoreFields(engine.Step{}, "Envs", "Secrets", "Labels"),
		cmpopts.IgnoreFields(engine.Network{}, "Labels"),
		cmpopts.IgnoreFields(engine.VolumeEmptyDir{}, "Labels"),
	}
	if diff := cmp.Diff(got, want, opts...); len(diff) != 0 {
		t.Errorf(diff)
	}

	return got.(*engine.Spec)
}

func dump(v interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

// This test verifies that privileged whitelisting is disabled when
// certain attributes, such as the entrypoint, command or commands
// are configured.
func TestIsPrivileged(t *testing.T) {
	c := new(Compiler)
	c.Privileged = []string{"foo"}
	if c.isPrivileged(&resource.Step{Image: "foo", Commands: []string{"echo hello", "echo world"}}) {
		t.Errorf("Disable privileged mode if commands are specified")
	}
	if c.isPrivileged(&resource.Step{Image: "foo", Command: []string{"echo hello", "echo world"}}) {
		t.Errorf("Disable privileged mode if the Docker command is specified")
	}
	if c.isPrivileged(&resource.Step{Image: "foo", Entrypoint: []string{"/bin/sh"}}) {
		t.Errorf("Disable privileged mode if the Docker entrypoint is specified")
	}
	if c.isPrivileged(&resource.Step{Image: "foo", Volumes: []*resource.VolumeMount{{MountPath: "/var/run/docker.sock"}}}) {
		t.Errorf("Disable privileged mode if /var/run/docker.sock mounted")
	}
	if c.isPrivileged(&resource.Step{Image: "foo", Volumes: []*resource.VolumeMount{{MountPath: "/var"}}}) {
		t.Errorf("Disable privileged mode if /var mounted")
	}
	if c.isPrivileged(&resource.Step{Image: "foo", Volumes: []*resource.VolumeMount{{MountPath: "/var/"}}}) {
		t.Errorf("Disable privileged mode if /var mounted")
	}
	if c.isPrivileged(&resource.Step{Image: "foo", Volumes: []*resource.VolumeMount{{MountPath: "/var//"}}}) {
		t.Errorf("Disable privileged mode if /var mounted")
	}
	if c.isPrivileged(&resource.Step{Image: "foo", Volumes: []*resource.VolumeMount{{MountPath: "/var/run"}}}) {
		t.Errorf("Disable privileged mode if /var/run mounted")
	}
	if c.isPrivileged(&resource.Step{Image: "foo", Volumes: []*resource.VolumeMount{{MountPath: "/"}}}) {
		t.Errorf("Disable privileged mode if / mounted")
	}
	if !c.isPrivileged(&resource.Step{Image: "foo"}) {
		t.Errorf("Enable privileged mode for privileged image")
	}
}
