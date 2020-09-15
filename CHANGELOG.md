The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
