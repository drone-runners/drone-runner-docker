package delegate

import (
	"github.com/dchest/uniuri"
	"github.com/drone-runners/drone-runner-docker/engine"
	"github.com/drone/runner-go/pipeline/runtime"
)

// TODO: Should be moved to compiler package

// TODO: random function from the compiler should be used instead
var random = func() string {
	return "drone-" + uniuri.NewLen(20)
}

func CompileDelegateStage() (runtime.Spec, error) {
	volumeID := random()
	networkID := random()

	vol := engine.Volume{
		EmptyDir: nil,
		HostPath: &engine.VolumeHostPath{
			ID:   volumeID,
			Name: "_workspace",
			//Path: "/home/tp/workspace/drone-runner-docker",
			Path: "/Users/markogacesa/volume",
			Labels: map[string]string{
				"io.drone.ttl": "1h0m0s"},
			ReadOnly: false,
		},
	}
	vols := []*engine.Volume{&vol}

	speccy := &engine.Spec{
		Network: engine.Network{
			ID: networkID,
			Labels: map[string]string{
				"io.drone.ttl": "1h0m0s",
			},
			Options: nil,
		},
		Volumes: vols,
	}

	return speccy, nil
}
