---
date: 2000-01-01T00:00:00+00:00
title: Plugins
author: bradrydzewski
weight: 5
toc: true
description: |
  Configure re-usable pipeline steps using plugins.
---

Plugins are docker containers that encapsulate commands, and can be shared and re-used in your pipeline. Use plugins to build and publish artifacts, send notifications, and more.

Example Slack plugin:

{{< highlight text "linenos=table,hl_lines=12-15" >}}
kind: pipeline
type: docker
name: default

steps:
- name: build
  image: node
  commands:
  - npm install
  - npm test

- name: notify
  image: plugins/slack
  settings:
    webhook: https://hooks.slack.com/services/...
{{< / highlight >}}

The great thing about plugins is they are just Docker containers. This means __anyone__ can easily encapsulate logic, bundle as a Docker image, and share with your organization or with the broader community.

# Secrets

Drone provides the ability to source any configuration parameter from a named secret using the `from_secret` syntax.

Example Slack plugin using secrets:

{{< highlight text "linenos=table,linenostart=5,hl_lines=11-12" >}}
steps:
- name: build
  image: node
  commands:
  - npm install
  - npm test

- name: notify
  image: plugins/slack
  settings:
    webhook:
      from_secret: webhook
{{< / highlight >}}

Example NPM plugin using secrets:

{{< highlight text "linenos=table,linenostart=5,hl_lines=11-14" >}}
steps:
- name: build
  image: node
  commands:
  - npm install
  - npm test

- name: publish
  image: plugins/npm
  settings:
    username:
      from_secret: username
    password:
      from_secret: password
{{< / highlight >}}

# Registry

The community maintains a public registry of Open Source plugins. You can browse the plugin registry at [plugins.drone.io](http://plugins.drone.io).