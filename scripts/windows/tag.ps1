# this script is used by the continuous integration server to
# build and publish the docker image for a tagged revsision.

$env:GOOS=windows
$env:GOARCH=amd64
$env:CGO_ENABLED=0
$env:VERSION=1903

# define the image tags
$env:IMAGE_PATCH=drone/drone-runner-docker:$env:DRONE_SEMVER_SHORT-windows-$env:VERSION-amd64
$env:IMAGE_MAJOR=drone/drone-runner-docker:$env:DRONE_SEMVER_MAJOR-windows-$env:VERSION-amd64
$env:IMAGE_MINOR=drone/drone-runner-docker:$env:DRONE_SEMVER_MAJOR.$env:DRONE_SEMVER_MINOR-windows-$env:VERSION-amd64

# build the binary
go build -o release/windows/amd64/drone-runner-docker.exe

# authenticate with the docker registry
docker login -u $env:USERNAME -p $env:PASSWORD

# build and tag the docker images
docker build -f docker/Dockerfile.windows.$env:VERSION -t $env:IMAGE_PATCH .
docker tag $env:IMAGE_PATCH $env:IMAGE_MAJOR
docker tag $env:IMAGE_PATCH $env:IMAGE_MINOR

# publish the docker images
docker push $env:IMAGE_MAJOR
docker push $env:IMAGE_MINOR
docker push $env:IMAGE_PATCH

# remove images after from local cache
docker rmi $env:IMAGE_MAJOR
docker rmi $env:IMAGE_MINOR
docker rmi $env:IMAGE_PATCH
