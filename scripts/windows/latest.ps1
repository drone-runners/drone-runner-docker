# this script is used by the continuous integration server to
# build and publish the docker image for a commit to master.

$env:GOOS="windows"
$env:GOARCH="amd64"
$env:CGO_ENABLED="0"

if (-not (Test-Path env:VERSION)) {
    $env:VERSION="1809"
}

echo $env:GOOS
echo $env:GOARCH
echo $env:VERSION

# build the binary
go build -o release/windows/amd64/drone-runner-docker.exe

# build and publish the docker image
docker login -u $env:USERNAME -p $env:PASSWORD
docker build -f docker/Dockerfile.windows.amd64.$env:VERSION -t drone/drone-runner-docker:windows-$env:VERSION-amd64 .
docker push drone/drone-runner-docker:windows-$env:VERSION-amd64

# remove images from local cache
docker rmi drone/drone-runner-docker:windows-$env:VERSION-amd64
