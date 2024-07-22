// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package command

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/drone-runners/drone-runner-docker/command/internal"
	"github.com/drone-runners/drone-runner-docker/engine/compiler"
	"github.com/drone-runners/drone-runner-docker/engine/linter"
	"github.com/drone-runners/drone-runner-docker/engine/resource"
	"github.com/drone/envsubst"
	"github.com/drone/runner-go/environ"
	"github.com/drone/runner-go/environ/provider"
	"github.com/drone/runner-go/manifest"
	"github.com/drone/runner-go/pipeline/runtime"
	"github.com/drone/runner-go/registry"
	"github.com/drone/runner-go/secret"

	harness "github.com/bradrydzewski/spec/yaml"
	"github.com/bradrydzewski/spec/yaml/utils/expand"
	"github.com/bradrydzewski/spec/yaml/utils/normalize"
	"github.com/bradrydzewski/spec/yaml/utils/resolver"
	compiler3 "github.com/drone-runners/drone-runner-docker/engine3/compiler"
	"github.com/drone-runners/drone-runner-docker/engine3/inputs"
	"github.com/drone-runners/drone-runner-docker/engine3/script"

	"gopkg.in/alecthomas/kingpin.v2"
)

type compileCommand struct {
	*internal.Flags

	Source     string
	Privileged []string
	Networks   []string
	Volumes    map[string]string
	Environ    map[string]string
	Labels     map[string]string
	Secrets    map[string]string
	Resources  compiler.Resources
	Templates  string
	Tmate      compiler.Tmate
	Clone      bool
	Config     string
}

func (c *compileCommand) run(pctx *kingpin.ParseContext) error {
	rawsource, err := ioutil.ReadFile(c.Source)
	if err != nil {
		return err
	}

	// if using the v1 yaml, use the new v1 run function
	if regexp.MustCompilePOSIX(`^pipeline:`).Match(rawsource) {
		return c.runv1(pctx)
	}

	envs := environ.Combine(
		c.Environ,
		environ.System(c.System),
		environ.Repo(c.Repo),
		environ.Build(c.Build),
		environ.Stage(c.Stage),
		environ.Link(c.Repo, c.Build, c.System),
		c.Build.Params,
	)

	// string substitution function ensures that string
	// replacement variables are escaped and quoted if they
	// contain newlines.
	subf := func(k string) string {
		v := envs[k]
		if strings.Contains(v, "\n") {
			v = fmt.Sprintf("%q", v)
		}
		return v
	}

	// evaluates string replacement expressions and returns an
	// update configuration.
	config, err := envsubst.Eval(string(rawsource), subf)
	if err != nil {
		return err
	}

	// parse and lint the configuration
	manifest, err := manifest.ParseString(config)
	if err != nil {
		return err
	}

	// a configuration can contain multiple pipelines.
	// get a specific pipeline resource for execution.
	resource, err := resource.Lookup(c.Stage.Name, manifest)
	if err != nil {
		return err
	}

	// lint the pipeline and return an error if any
	// linting rules are broken
	lint := linter.New()
	err = lint.Lint(resource, c.Repo)
	if err != nil {
		return err
	}

	// compile the pipeline to an intermediate representation.
	comp := &compiler.Compiler{
		Environ:    provider.Static(c.Environ),
		Labels:     c.Labels,
		Resources:  c.Resources,
		Tmate:      c.Tmate,
		Privileged: append(c.Privileged, compiler.Privileged...),
		Networks:   c.Networks,
		Volumes:    c.Volumes,
		Secret:     secret.StaticVars(c.Secrets),
		Registry: registry.Combine(
			registry.File(c.Config),
		),
	}

	// when running a build locally cloning is always
	// disabled in favor of mounting the source code
	// from the current working directory.
	if c.Clone == false {
		comp.Mount, _ = os.Getwd()
	}

	args := runtime.CompilerArgs{
		Pipeline: resource,
		Manifest: manifest,
		Build:    c.Build,
		Netrc:    c.Netrc,
		Repo:     c.Repo,
		Stage:    c.Stage,
		System:   c.System,
		Secret:   secret.StaticVars(c.Secrets),
	}
	spec := comp.Compile(nocontext, args)

	// encode the pipeline in json format and print to the
	// console for inspection.
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(spec)
	return nil
}

func (c *compileCommand) runv1(*kingpin.ParseContext) error {

	config, err := harness.ParseFile(c.Source)
	if err != nil {
		return err
	}

	//
	// expand matrix stages and steps
	//

	expand.Expand(config)

	//
	// expand templates and plugins
	//

	// expand templates and plugins
	resolve := func(name string) (*harness.Template, error) {
		path := filepath.Join(c.Templates, name+".yaml")
		schema, err := harness.ParseFile(path)
		if err != nil {
			return nil, err
		}
		return schema.Template, nil
	}
	if err := resolver.Resolve(config, resolve); err != nil {
		return err
	}

	//
	// expand name and identifier fields.
	// other fields are expanded at execution time.
	//
	// NOTE this is a huge hack

	inputParams := map[string]interface{}{}
	inputParams["repo"] = inputs.Repo(c.Repo)
	inputParams["build"] = inputs.Build(c.Build)
	if config.Pipeline.Inputs != nil {
		inputParams = inputs.Inputs(config.Pipeline.Inputs, c.Build.Params)
	}
	script.ExpandConfig(config, inputParams)

	// normalize the configuration to ensure
	// all steps have an identifier
	normalize.Normalize(config)

	// FIXME: find a better way to do this
	// HACK get the default stage name
	if c.Stage.Name == "" || c.Stage.Name == "default" {
		// use the normalized id to refer to the stage
		// FIXME: this won't work in some cases, and will cause problems
		c.Stage.Name = config.Pipeline.Stages[0].Id
	}

	// compile the pipeline to an intermediate representation.
	comp := &compiler3.CompilerImpl{
		Environ: provider.Static(c.Environ),
		Labels:  c.Labels,
		// TODO re-add
		// Resources:  c.Resources,
		// Tmate:      c.Tmate,
		Privileged: append(c.Privileged, compiler.Privileged...),
		Networks:   c.Networks,
		Volumes:    c.Volumes,
		Secret:     secret.StaticVars(c.Secrets),
		Registry: registry.Combine(
			registry.File(c.Config),
		),
	}

	// when running a build locally cloning is always
	// disabled in favor of mounting the source code
	// from the current working directory.
	if c.Clone == false {
		comp.Mount, _ = os.Getwd()
	}

	args := compiler3.Args{
		Config: config,
		Build:  c.Build,
		Netrc:  c.Netrc,
		Repo:   c.Repo,
		Stage:  c.Stage,
		System: c.System,
		Secret: secret.StaticVars(c.Secrets),
	}
	spec, err := comp.Compile(nocontext, args)
	if err != nil {
		return err
	}

	// encode the pipeline in json format and print to the
	// console for inspection.
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(spec)
	return nil
}

func registerCompile(app *kingpin.Application) {
	c := new(compileCommand)
	c.Environ = map[string]string{}
	c.Secrets = map[string]string{}
	c.Labels = map[string]string{}
	c.Volumes = map[string]string{}

	cmd := app.Command("compile", "compile the yaml file").
		Action(c.run)

	cmd.Flag("source", "source file location").
		Default(".drone.yml").
		StringVar(&c.Source)

	cmd.Flag("clone", "enable cloning").
		BoolVar(&c.Clone)

	cmd.Flag("secrets", "secret parameters").
		StringMapVar(&c.Secrets)

	cmd.Flag("templates", "template directory").
		StringVar(&c.Templates)

	cmd.Flag("environ", "environment variables").
		StringMapVar(&c.Environ)

	cmd.Flag("labels", "container labels").
		StringMapVar(&c.Labels)

	cmd.Flag("networks", "container networks").
		StringsVar(&c.Networks)

	cmd.Flag("volumes", "container volumes").
		StringMapVar(&c.Volumes)

	cmd.Flag("privileged", "privileged docker images").
		StringsVar(&c.Privileged)

	cmd.Flag("cpu-period", "container cpu period").
		Int64Var(&c.Resources.CPUPeriod)

	cmd.Flag("cpu-quota", "container cpu quota").
		Int64Var(&c.Resources.CPUQuota)

	cmd.Flag("cpu-set", "container cpu set").
		StringsVar(&c.Resources.CPUSet)

	cmd.Flag("cpu-shares", "container cpu shares").
		Int64Var(&c.Resources.CPUShares)

	cmd.Flag("memory", "container memory limit").
		Int64Var(&c.Resources.Memory)

	cmd.Flag("memory-swap", "container memory swap limit").
		Int64Var(&c.Resources.MemorySwap)

	cmd.Flag("docker-config", "path to the docker config file").
		StringVar(&c.Config)

	cmd.Flag("tmate-image", "tmate docker image").
		Default("drone/drone-runner-docker:1").
		StringVar(&c.Tmate.Image)

	cmd.Flag("tmate-enabled", "tmate enabled").
		BoolVar(&c.Tmate.Enabled)

	cmd.Flag("tmate-server-host", "tmate server host").
		StringVar(&c.Tmate.Server)

	cmd.Flag("tmate-server-port", "tmate server port").
		StringVar(&c.Tmate.Port)

	cmd.Flag("tmate-server-rsa-fingerprint", "tmate server rsa fingerprint").
		StringVar(&c.Tmate.RSA)

	cmd.Flag("tmate-server-ed25519-fingerprint", "tmate server rsa fingerprint").
		StringVar(&c.Tmate.ED25519)

	cmd.Flag("tmate-authorized-keys", "tmate authorized keys").
		StringVar(&c.Tmate.AuthorizedKeys)

	// shared pipeline flags
	c.Flags = internal.ParseFlags(cmd)
}
