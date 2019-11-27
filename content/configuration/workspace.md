---
date: 2000-01-01T00:00:00+00:00
title: Workspace
author: bradrydzewski
weight: 2
toc: false
description: |
  Describes the pipeline workspace and directory structure.
---

Drone automatically creates a temporary volume, known as your workspace, where it clones your repository. The workspace is the current working directory for each step in your pipeline.

Because the workspace is a volume, filesystem changes are persisted between pipeline steps. In other words, individual steps can communicate and share state using the filesystem.

Workspace path inside your pipeline containers:

```
/drone/src
```

{{< alert "warn" >}}
Note the workspace volume is ephemeral. It is created when the pipeline starts and destroyed after the pipeline completes.
{{< / alert >}}

# Customizing the Workspace

You can customize the workspace directory by defining the `workspace` section in your yaml. Here is a basic example:

{{< highlight text "linenos=table,linenostart=1,hl_lines=5-7" >}}
kind: pipeline
type: docker
name: default

workspace:
  base: /go
  path: src/github.com/octocat/hello-world

steps:
- name: backend
  image: golang:latest
  commands:
  - go get
  - go test

- name: frontend
  image: node:latest
  commands:
  - npm install
  - npm run tests
{{< / highlight >}}

This would be equivalent to the following docker commands:

```text
$ docker volume create my-named-volume
$ docker run --volume=my-named-volume:/go golang:latest
$ docker run --volume=my-named-volume:/go node:latest
```

The `base` attribute defines a shared base volume available to all pipeline steps. This ensures your source code, dependencies and compiled binaries are persisted and shared between steps.

{{< highlight text "linenos=table,linenostart=1,hl_lines=6" >}}
kind: pipeline
type: docker
name: default

workspace:
  base: /go
  path: src/github.com/octocat/hello-world
{{< / highlight >}}

The `path` attribute defines the working directory of your build. This is where your code is cloned and will be the default working directory of every step in your build process. The path must be relative and is combined with your base path.

{{< highlight text "linenos=table,linenostart=1,hl_lines=7" >}}
kind: pipeline
type: docker
name: default

workspace:
  base: /go
  path: src/github.com/octocat/hello-world
{{< / highlight >}}

Example clone path when using the above configuration:

```text
$ git clone https://github.com/octocat/hello-world \
  /go/src/github.com/octocat/hello-world
```
