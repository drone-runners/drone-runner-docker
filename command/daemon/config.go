// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package daemon

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Config stores the system configuration.
type Config struct {
	Debug bool `envconfig:"DRONE_DEBUG"`
	Trace bool `envconfig:"DRONE_TRACE"`

	Client struct {
		Address    string `ignored:"true"`
		Proto      string `envconfig:"DRONE_RPC_PROTO"  default:"http"`
		Host       string `envconfig:"DRONE_RPC_HOST"   required:"true"`
		Secret     string `envconfig:"DRONE_RPC_SECRET" required:"true"`
		SkipVerify bool   `envconfig:"DRONE_RPC_SKIP_VERIFY"`
		Dump       bool   `envconfig:"DRONE_RPC_DUMP_HTTP"`
		DumpBody   bool   `envconfig:"DRONE_RPC_DUMP_HTTP_BODY"`
	}

	Dashboard struct {
		Disabled bool   `envconfig:"DRONE_UI_DISABLE"`
		Username string `envconfig:"DRONE_UI_USERNAME"`
		Password string `envconfig:"DRONE_UI_PASSWORD"`
		Realm    string `envconfig:"DRONE_UI_REALM" default:"MyRealm"`
	}

	Server struct {
		Port  string `envconfig:"DRONE_HTTP_BIND" default:":3000"`
		Proto string `envconfig:"DRONE_HTTP_PROTO"`
		Host  string `envconfig:"DRONE_HTTP_HOST"`
		Acme  bool   `envconfig:"DRONE_HTTP_ACME"`
	}

	Runner struct {
		Name              string            `envconfig:"DRONE_RUNNER_NAME"`
		Capacity          int               `envconfig:"DRONE_RUNNER_CAPACITY" default:"2"`
		Procs             int64             `envconfig:"DRONE_RUNNER_MAX_PROCS"`
		Environ           map[string]string `envconfig:"DRONE_RUNNER_ENVIRON"`
		EnvFile           string            `envconfig:"DRONE_RUNNER_ENV_FILE"`
		Secrets           map[string]string `envconfig:"DRONE_RUNNER_SECRETS"`
		Labels            map[string]string `envconfig:"DRONE_RUNNER_LABELS"`
		Volumes           map[string]string `envconfig:"DRONE_RUNNER_VOLUMES"`
		Devices           []string          `envconfig:"DRONE_RUNNER_DEVICES"`
		Networks          []string          `envconfig:"DRONE_RUNNER_NETWORKS"`
		NetworkEnableIPv6 bool              `envconfig:"DRONE_RUNNER_NETWORK_ENABLE_IPV6" default:"false"`
		NetworkOpts       map[string]string `envconfig:"DRONE_RUNNER_NETWORK_OPTS"`
		Privileged        []string          `envconfig:"DRONE_RUNNER_PRIVILEGED_IMAGES"`
		Clone             string            `envconfig:"DRONE_RUNNER_CLONE_IMAGE"`
	}

	Platform struct {
		OS      string `envconfig:"DRONE_PLATFORM_OS"    default:"linux"`
		Arch    string `envconfig:"DRONE_PLATFORM_ARCH"  default:"amd64"`
		Kernel  string `envconfig:"DRONE_PLATFORM_KERNEL"`
		Variant string `envconfig:"DRONE_PLATFORM_VARIANT"`
	}

	Limit struct {
		Repos   []string `envconfig:"DRONE_LIMIT_REPOS"`
		Events  []string `envconfig:"DRONE_LIMIT_EVENTS"`
		Trusted bool     `envconfig:"DRONE_LIMIT_TRUSTED"`
	}

	Resources struct {
		Memory     int64    `envconfig:"DRONE_MEMORY_LIMIT"`
		MemorySwap int64    `envconfig:"DRONE_MEMORY_SWAP_LIMIT"`
		CPUQuota   int64    `envconfig:"DRONE_CPU_QUOTA"`
		CPUPeriod  int64    `envconfig:"DRONE_CPU_PERIOD"`
		CPUShares  int64    `envconfig:"DRONE_CPU_SHARES"`
		CPUSet     []string `envconfig:"DRONE_CPU_SET"`
		ShmSize    int64    `envconfig:"DRONE_SHM_SIZE"`
	}

	Environ struct {
		Endpoint   string `envconfig:"DRONE_ENV_PLUGIN_ENDPOINT"`
		Token      string `envconfig:"DRONE_ENV_PLUGIN_TOKEN"`
		SkipVerify bool   `envconfig:"DRONE_ENV_PLUGIN_SKIP_VERIFY"`
	}

	Secret struct {
		Endpoint   string `envconfig:"DRONE_SECRET_PLUGIN_ENDPOINT"`
		Token      string `envconfig:"DRONE_SECRET_PLUGIN_TOKEN"`
		SkipVerify bool   `envconfig:"DRONE_SECRET_PLUGIN_SKIP_VERIFY"`
	}

	Netrc struct {
		CloneOnly bool `envconfig:"DRONE_NETRC_CLONE_ONLY"`
	}

	Registry struct {
		Endpoint   string `envconfig:"DRONE_REGISTRY_PLUGIN_ENDPOINT"`
		Token      string `envconfig:"DRONE_REGISTRY_PLUGIN_TOKEN"`
		SkipVerify bool   `envconfig:"DRONE_REGISTRY_PLUGIN_SKIP_VERIFY"`
	}

	Docker struct {
		Config string `envconfig:"DRONE_DOCKER_CONFIG"`
		Stream bool   `envconfig:"DRONE_DOCKER_STREAM_PULL" default:"true"`
	}

	Tmate struct {
		Enabled        bool   `envconfig:"DRONE_TMATE_ENABLED" default:"false"`
		Image          string `envconfig:"DRONE_TMATE_IMAGE"   default:"drone/drone-runner-docker:1"`
		Server         string `envconfig:"DRONE_TMATE_HOST"`
		Port           string `envconfig:"DRONE_TMATE_PORT"`
		RSA            string `envconfig:"DRONE_TMATE_FINGERPRINT_RSA"`
		ED25519        string `envconfig:"DRONE_TMATE_FINGERPRINT_ED25519"`
		AuthorizedKeys string `envconfig:"DRONE_TMATE_AUTHORIZED_KEYS"`
	}
}

// legacy environment variables. the key is the legacy
// variable name, and the value is the new variable name.
var legacy = map[string]string{
	"DRONE_MACHINE":                "DRONE_RUNNER_NAME",
	"DRONE_RUNNER_OS":              "DRONE_PLATFORM_OS",
	"DRONE_RUNNER_ARCH":            "DRONE_PLATFORM_ARCH",
	"DRONE_RUNNER_KERNEL":          "DRONE_PLATFORM_KERNEL",
	"DRONE_RUNNER_VARIANT":         "DRONE_PLATFORM_VARIANT",
	"DRONE_REGISTRY_ENDPOINT":      "DRONE_REGISTRY_PLUGIN_ENDPOINT",
	"DRONE_REGISTRY_SECRET":        "DRONE_REGISTRY_PLUGIN_TOKEN",
	"DRONE_REGISTRY_PLUGIN_SECRET": "DRONE_REGISTRY_PLUGIN_TOKEN",
	"DRONE_REGISTRY_SKIP_VERIFY":   "DRONE_REGISTRY_PLUGIN_SKIP_VERIFY",
	"DRONE_SECRET_ENDPOINT":        "DRONE_SECRET_PLUGIN_ENDPOINT",
	"DRONE_SECRET_SECRET":          "DRONE_SECRET_PLUGIN_TOKEN",
	"DRONE_SECRET_SKIP_VERIFY":     "DRONE_SECRET_PLUGIN_SKIP_VERIFY",
	"DRONE_LIMIT_MEM_SWAP":         "DRONE_MEMORY_SWAP_LIMIT",
	"DRONE_LIMIT_MEM":              "DRONE_MEMORY_LIMIT",
	"DRONE_LIMIT_CPU_QUOTA":        "DRONE_CPU_QUOTA",
	"DRONE_LIMIT_CPU_SHARES":       "DRONE_CPU_SHARES",
	"DRONE_LIMIT_CPU_SET":          "DRONE_CPU_SET",
	"DRONE_LOGS_DEBUG":             "DRONE_DEBUG",
	"DRONE_LOGS_TRACE":             "DRONE_TRACE",
	"DRONE_SERVER_PROTO":           "DRONE_HTTP_PROTO",
	"DRONE_SERVER_HOST":            "DRONE_HTTP_HOST",
	"DRONE_SERVER_PORT":            "DRONE_HTTP_BIND",
	"DRONE_SERVER_ACME":            "DRONE_HTTP_ACME",
}

func fromEnviron() (Config, error) {
	// loop through legacy environment variable and, if set
	// rewrite to the new variable name.
	for k, v := range legacy {
		if s, ok := os.LookupEnv(k); ok {
			os.Setenv(v, s)
		}
	}

	var config Config
	err := envconfig.Process("", &config)
	if err != nil {
		return config, err
	}
	if config.Runner.Environ == nil {
		config.Runner.Environ = map[string]string{}
	}
	if config.Runner.Name == "" {
		config.Runner.Name, _ = os.Hostname()
	}
	if config.Dashboard.Password == "" {
		config.Dashboard.Disabled = true
	}
	config.Client.Address = fmt.Sprintf(
		"%s://%s",
		config.Client.Proto,
		config.Client.Host,
	)

	// environment variables can be sourced from a separate
	// file. These variables are loaded and appended to the
	// environment list.
	if file := config.Runner.EnvFile; file != "" {
		envs, err := godotenv.Read(file)
		if err != nil {
			return config, err
		}
		for k, v := range envs {
			config.Runner.Environ[k] = v
		}
	}

	return config, nil
}
