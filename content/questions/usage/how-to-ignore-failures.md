---
date: 2000-01-01T00:00:00+00:00
title: How to ignore steps that fail?
author: bradrydzewski
weight: 1
draft: true
---


```
kind: pipeline
type: docker
name: default

steps:
- name: foo
  failure: ignore
  commands:
  - echo foo
  - exit 1

- name: bar
  commands:
  - echo bar
```