# go version
go = "golang:1.13"

# docker repository
repo = "drone/drone-runner-docker"

def main(ctx):
    return [
        pipeline_linux(),
        pipeline_windows("1903"),
        pipeline_windows("1809"),
        pipeline_windows("2022"),
        pipeline_manifest(),
        # manifest
    ]


def pipeline_linux():
    return  {
        "kind": "pipeline",
        "type": "docker",
        "name": "linux",
        "steps": [
            test,
            build,
            publish("arm"),
            publish("arm64"),
            publish("amd64"),
        ]
    }

def pipeline_windows(version):
    return {
        "kind": "pipeline",
        "type": "ssh",
        "name": "windows_%s" % version,
        "server": {

        },
        "platform": {
            "os": "windows",
        },
        "steps": [
            "sh scripts/ci_%s.ps1" % version,
        ],
    }

def pipeline_manifest():
    return [
        "kind": "pipeline",
        "type": "docker",
        "name": "linux",
        "steps": [
            {

            }
        ]
    ]

# publish creates a docker publish step.
def publish(arch):
    return {
        "name": "publish_%s" % arch,
        "image": "plugins/docker",
        "pull": "if-not-exists",
        "settings": {
            "auto_tag": "true",
            "auto_tag_suffix": "linux-%s" % arch,
            "dockerfile": "docker/Dockerfile.linux.%s" % arch,
            "repo": repo,
        },
        "when": {
            "event": [ "push", "tag" ]
        }
    }

# test defines a test step that downloads
# dependencies and tests the packages.
test = {
    "name": "test",
    "image": go,
    "volumes": mounts,
    "commands": [
        "go test -v ./...",
    ],
}

# build defines a build step that compiles
# the binaries.
build = {
    "name": "build",
    "image": go,
    "volumes": mounts,
    "commands": [
        "sh scripts/build.sh",
    ],
    "when": {
        "event": [ "push", "tag" ]
    }
}

# mount points shared by all steps.
mounts = [{
    "name": "go",
    "path": "/go",
}]
