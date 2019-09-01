# Kubeval

`kubeval` is a tool for validating a Kubernetes YAML or JSON configuration file.
It does so using schemas generated from the Kubernetes OpenAPI specification, and
therefore can validate schemas for multiple versions of Kubernetes.

[![CircleCI](https://circleci.com/gh/instrumenta/kubeval.svg?style=svg)](https://circleci.com/gh/instrumenta/kubeval)
[![Go Report
Card](https://goreportcard.com/badge/github.com/instrumenta/kubeval)](https://goreportcard.com/report/github.com/instrumenta/kubeval)
[![GoDoc](https://godoc.org/github.com/instrumenta/kubeval?status.svg)](https://godoc.org/github.com/instrumenta/kubeval)


```
$ kubeval my-invalid-rc.yaml
WARN - fixtures/my-invalid-rc.yaml contains an invalid ReplicationController - spec.replicas: Invalid type. Expected: [integer,null], given: string
$ echo $?
1
```


For full usage and installation instructions see [kubeval.instrumenta.dev](https://kubeval.instrumenta.dev/).
