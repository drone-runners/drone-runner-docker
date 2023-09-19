module github.com/drone-runners/drone-runner-docker

go 1.20

replace github.com/docker/docker => github.com/docker/engine v17.12.0-ce-rc1.0.20200309214505-aa6a9891b09c+incompatible

require (
	github.com/antonmedv/expr v1.15.2
	github.com/buildkite/yaml v2.1.0+incompatible
	github.com/dchest/uniuri v0.0.0-20160212164326-8902c56451e9
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v0.0.0-00010101000000-000000000000
	github.com/drone/drone-go v1.7.1
	github.com/drone/envsubst v1.0.3
	github.com/drone/runner-go v1.12.0
	github.com/drone/signal v1.0.0
	github.com/drone/spec v0.0.0-20230918220055-b61fc77555aa
	github.com/ghodss/yaml v1.0.0
	github.com/google/go-cmp v0.5.9
	github.com/hashicorp/go-multierror v1.0.0
	github.com/joho/godotenv v1.3.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/mattn/go-isatty v0.0.8
	github.com/natessilva/dag v0.0.0-20180124060714-7194b8dcc5c4
	github.com/sirupsen/logrus v1.4.2
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
)

require (
	github.com/99designs/basicauth-go v0.0.0-20160802081356-2a93ba0f464d // indirect
	github.com/99designs/httpsignatures-go v0.0.0-20170731043157-88528bf4ca7e // indirect
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/Microsoft/go-winio v0.4.11 // indirect
	github.com/alecthomas/template v0.0.0-20160405071501-a0175ee3bccc // indirect
	github.com/alecthomas/units v0.0.0-20151022065526-2efee857e7cf // indirect
	github.com/bmatcuk/doublestar v1.3.4 // indirect
	github.com/containerd/containerd v1.3.4 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/docker/go-connections v0.3.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/gogo/protobuf v0.0.0-20170307180453-100ba4e88506 // indirect
	github.com/golang/protobuf v1.4.1 // indirect
	github.com/gorilla/mux v1.7.4 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.1 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/pkg/errors v0.8.1 // indirect
	golang.org/x/crypto v0.0.0-20190621222207-cc06ce4a13d4 // indirect
	golang.org/x/net v0.0.0-20190404232315-eb5bcb51f2a3 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	golang.org/x/time v0.0.0-20181108054448-85acf8d2951c // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013 // indirect
	google.golang.org/grpc v1.29.1 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gotest.tools v2.2.0+incompatible // indirect
)
