# Changelog

## [v1.8.3](https://github.com/drone-runners/drone-runner-docker/tree/v1.8.3) (2023-01-06)

[Full Changelog](https://github.com/drone-runners/drone-runner-docker/compare/v1.8.2...v1.8.3)

**Fixed bugs:**

- \(dron-434\) set platform in docker build step [\#57](https://github.com/drone-runners/drone-runner-docker/pull/57) ([tphoney](https://github.com/tphoney))
- Revert "Dockerfile: Add target architecture to Docker images" [\#56](https://github.com/drone-runners/drone-runner-docker/pull/56) ([tphoney](https://github.com/tphoney))
- Dockerfile: Add target architecture to Docker images [\#54](https://github.com/drone-runners/drone-runner-docker/pull/54) ([jnohlgard](https://github.com/jnohlgard))

**Merged pull requests:**

- \(maint\) move to harness.drone.io [\#59](https://github.com/drone-runners/drone-runner-docker/pull/59) ([tphoney](https://github.com/tphoney))

## [v1.8.2](https://github.com/drone-runners/drone-runner-docker/tree/v1.8.2) (2022-06-14)

[Full Changelog](https://github.com/drone-runners/drone-runner-docker/compare/v1.8.1...v1.8.2)

**Implemented enhancements:**

- Add retries option to the clone step [\#44](https://github.com/drone-runners/drone-runner-docker/pull/44) ([julienduchesne](https://github.com/julienduchesne))

**Merged pull requests:**

- release prep for v1.8.2 [\#49](https://github.com/drone-runners/drone-runner-docker/pull/49) ([eoinmcafee00](https://github.com/eoinmcafee00))

## [v1.8.1](https://github.com/drone-runners/drone-runner-docker/tree/v1.8.1) (2022-04-19)

[Full Changelog](https://github.com/drone-runners/drone-runner-docker/compare/v1.8.0...v1.8.1)

**Fixed bugs:**

- \(dron-254\) handle wait for log transfer [\#46](https://github.com/drone-runners/drone-runner-docker/pull/46) ([tphoney](https://github.com/tphoney))
- Fix 3 doc typos in compiler.go [\#45](https://github.com/drone-runners/drone-runner-docker/pull/45) ([mach6](https://github.com/mach6))
- feat\(engine\): Add debug logs for Docker.Destroy errors [\#20](https://github.com/drone-runners/drone-runner-docker/pull/20) ([jvrplmlmn](https://github.com/jvrplmlmn))

**Merged pull requests:**

- \(maint\) release prep v1.8.1 [\#47](https://github.com/drone-runners/drone-runner-docker/pull/47) ([tphoney](https://github.com/tphoney))

## [v1.8.0](https://github.com/drone-runners/drone-runner-docker/tree/v1.8.0) (2021-11-18)

[Full Changelog](https://github.com/drone-runners/drone-runner-docker/compare/v1.7.0...v1.8.0)

**Implemented enhancements:**

- create and store card data [\#41](https://github.com/drone-runners/drone-runner-docker/pull/41) ([eoinmcafee00](https://github.com/eoinmcafee00))

**Merged pull requests:**

- release prep for v1.8.0 [\#43](https://github.com/drone-runners/drone-runner-docker/pull/43) ([eoinmcafee00](https://github.com/eoinmcafee00))

## [v1.7.0](https://github.com/drone-runners/drone-runner-docker/tree/v1.7.0) (2021-11-02)

[Full Changelog](https://github.com/drone-runners/drone-runner-docker/compare/v1.6.3...v1.7.0)

**Implemented enhancements:**

- \(maint\) prep for v1.7.0 [\#42](https://github.com/drone-runners/drone-runner-docker/pull/42) ([tphoney](https://github.com/tphoney))
- Expose the authorized keys tmate feature [\#18](https://github.com/drone-runners/drone-runner-docker/pull/18) ([julienduchesne](https://github.com/julienduchesne))
- Support ppc64le [\#17](https://github.com/drone-runners/drone-runner-docker/pull/17) ([isuruf](https://github.com/isuruf))
- \(feat\) adding image field to step [\#14](https://github.com/drone-runners/drone-runner-docker/pull/14) ([tphoney](https://github.com/tphoney))
- check privileged image whitelist in service section [\#9](https://github.com/drone-runners/drone-runner-docker/pull/9) ([divialth](https://github.com/divialth))

**Fixed bugs:**

- \(maint\) bump go version in build to match go.mod [\#40](https://github.com/drone-runners/drone-runner-docker/pull/40) ([tphoney](https://github.com/tphoney))
- bump drone/runner-go to include trigger env var [\#32](https://github.com/drone-runners/drone-runner-docker/pull/32) ([willejs](https://github.com/willejs))
- Use correct name for the root depends\_on yaml key. [\#28](https://github.com/drone-runners/drone-runner-docker/pull/28) ([staffanselander](https://github.com/staffanselander))
- \(DRON-82\) update envsubst to prevent compiler panics [\#24](https://github.com/drone-runners/drone-runner-docker/pull/24) ([tphoney](https://github.com/tphoney))

**Merged pull requests:**

- Use IsRestrictedVolume from runner-go [\#26](https://github.com/drone-runners/drone-runner-docker/pull/26) ([marko-gacesa](https://github.com/marko-gacesa))
- chore: update runner-go to 1.8.0 [\#25](https://github.com/drone-runners/drone-runner-docker/pull/25) ([fgierlinger](https://github.com/fgierlinger))
- Test that inverse matches work for `DRONE_LIMIT_REPOS` [\#19](https://github.com/drone-runners/drone-runner-docker/pull/19) ([jvrplmlmn](https://github.com/jvrplmlmn))

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## 1.6.3
### Fixed
- use path prefix when evaluating restricted volume mounts. See [#ea74fa2](https://github.com/drone-runners/drone-runner-docker/commit/ea74fa2ba442eacb0812ad5983c305a16b6763bc).

## 1.6.2
### Added
- support for self-hosted tmate instances

## 1.6.1
### Changed
- restrict temporary volumes used with docker plugins
- restrict environment variables used with docker plugins

## 1.6.0
### Added
- experimental support for remote debugging with tmate, disabled by default

### Fixed
- exit code 78 not properly exiting early when pipeline has services (from runner-go)

## 1.5.3
### Fixed
- unexpected http code from server must always fail pipeline (from runner-go)

## 1.5.2
### Added
- trace logging for semaphore acquisition and release

### Fixed
- failure to acquire semaphore due to error should fail the pipeline
- failure to acquire semaphore due to context deadline should cancel the pipeline

## 1.5.1
### Fixed
- cancel a build should result in cancel status, not error status

## 1.5.0
### Added
- option to disable netrc for non-clone steps
- option to customize docker bridge networks

### Changed
- upgrade docker client

## 1.4.0
### Added
- support for windows 1909
- support for nomad runner execution

## 1.3.0
### Added
- support for setting default container shmsize

### Changed
- update environment extension protocol to version 2
- registry credentials stored in repository secrets take precedence over globals

### Fixed
- ignoring global memory limit and memory swap limit

### Added
- support for environment extension variable masking
- support for username/password in config.json files

## 1.2.1
### Added
- deployment id environment variable
- support for multi-line secret masking
- trace logging prints external registry details

### Fixed
- do not override user defined mem_limit and memswap_limit
- remove scheme when comparing image and registry hostnames

## 1.2.0
### Added
- support for mem_limit and memswap_limit

## 1.1.0
### Changed

- abstract polling and execution to runner-go library
- use trace level logging for context errors
- prefix docker resource names

### Fixed
- initialize environment map to prevent nil pointer

### Added
- support for global environment variable extension

## 1.0.1
### Fixed

- handle pipelines with missing names
- prevent mounting /run/drone directory

## 1.0.0
### Added

- ported docker pipelines to runner-go framework
- support for pipeline environment variables
- support for shm_size


\* *This Changelog was automatically generated by [github_changelog_generator](https://github.com/github-changelog-generator/github-changelog-generator)*
