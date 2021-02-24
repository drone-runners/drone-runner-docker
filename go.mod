module github.com/drone-runners/drone-runner-docker

go 1.12

replace github.com/docker/docker => github.com/docker/engine v17.12.0-ce-rc1.0.20200309214505-aa6a9891b09c+incompatible

require (
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/Microsoft/go-winio v0.4.11 // indirect
	github.com/buildkite/yaml v2.1.0+incompatible
	github.com/containerd/containerd v1.3.4 // indirect
	github.com/dchest/uniuri v0.0.0-20160212164326-8902c56451e9
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v0.0.0-00010101000000-000000000000
	github.com/docker/go-connections v0.3.0 // indirect
	github.com/drone/drone-go v1.6.0
	github.com/drone/envsubst v1.0.2
	github.com/drone/runner-go v1.7.0
	github.com/drone/signal v1.0.0
	github.com/ghodss/yaml v1.0.0
	github.com/google/go-cmp v0.4.0
	github.com/gorilla/mux v1.7.4 // indirect
	github.com/joho/godotenv v1.3.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/kr/pretty v0.1.0 // indirect
	github.com/mattn/go-isatty v0.0.8
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/prometheus/client_golang v1.9.0 // indirect
	github.com/sirupsen/logrus v1.6.0
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	google.golang.org/grpc v1.29.1 // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gotest.tools v2.2.0+incompatible // indirect
)
