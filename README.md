# Kubeval

`kubeval` is a tool for validating a Kubernetes YAML or JSON configuration file.
It can also be used as a library in other Go applications.

[![Build
Status](https://travis-ci.org/garethr/kubeval.svg)](https://travis-ci.org/garethr/kubeval)
[![Go Report
Card](https://goreportcard.com/badge/github.com/garethr/kubeval)](https://goreportcard.com/report/github.com/garethr/kubeval)
[![GoDoc](https://godoc.org/github.com/garethr/kubeval?status.svg)](https://godoc.org/github.com/garethr/kubeval)
[![Coverage
Status](https://coveralls.io/repos/github/garethr/kubeval/badge.svg?branch=master)](https://coveralls.io/github/garethr/kubeval?branch=master)

```
$ kubeval my-invalid-rc.yaml
The document my-invalid-rc.yaml contains an invalid ReplicationController
--> spec.replicas: Invalid type. Expected: integer, given: string
$ echo $?
1
```

Alternatively kubeval can also take input via `stdin` which can make using
it as part of an automated pipeline easier.

```
$ cat my-invalid-rc.yaml | kubeval
The document my-invalid-rc.yaml contains an invalid ReplicationController
--> spec.replicas: Invalid type. Expected: integer, given: string
$ echo $?
1
```


## Why?

* If you're writing Kubernetes configuration files by hand it is useful
  to check them for validity before applying them
* If you're distributing Kubernetes configuration files or examples it's
  handy to check them against multiple versions of Kubernetes
* If you're generating Kubernetes configurations using a tool like
  ksonnet or hand-rolled templating it's important to make sure the
  output is valid

I'd like to be able to address the above both locally when developing,
and also as a simple gate in a continuous integration system.

`kubectl` doesn't address the above needs in a few ways, importantly
validating with `kubectl` requires a Kubernetes cluster. If you want to
validate against multiple versions of Kubernetes, you'll need multiple
clusters. All of that for validating the structure of a data structure
stored in plain text makes for an unweild development environment.


## But how?

Kubernetes has strong definitions of what a Deployment, Pod, or
ReplicationController are. It exposes that information via an OpenAPI
based description. That description contains JSON Schema information for
the Kubernetes types. This tool uses those extracted schemas, published
at [garethr/kubernetes-json-schema](https://github.com/garethr/kubernetes-json-schema) and [garethr/openshift-json-schema](https://github.com/garethr/openshift-json-schema). See
those repositories and
[this blog post](https://www.morethanseven.net/2017/06/26/schemas-for-kubernetes-types/)
for the details.


## Installation

Tagged versions of `kubeval` are built by Travis and automatically
uploaded to GitHub. This means you should find `tar.gz` files under the
release tab. These should contain a single `kubeval` binary for platform
in the filename (ie. windows, linux, darwin). Either execute that binary
directly or place it on your path.

```
wget
https://github.com/garethr/kubeval/releases/download/0.6.0/kubeval-darwin-amd64.tar.gz
tar xf kubeval-darwin-amd64.tar.gz
cp kubeval /usr/local/bin
```

Windows users can download tar or zip files from the releases, or for [Chocolatey](https://chocolatey.org)
users you can install with:

```
choco install kubeval
```

For those on macOS using [Homebrew](https://brew.sh/) you can use the kubeval tap:

```
brew tap garethr/kubeval
brew install kubeval
```

`kubeval` is also published as a Docker image. So can be used as
follows:

```
$ docker run -it -v `pwd`/fixtures:/fixtures garethr/kubeval fixtures/*
Missing a kind key in /fixtures/blank.yaml
The document fixtures/int_or_string.yaml contains a valid Service
The document fixtures/int_or_string_false.yaml contains an invalid Deployment
--> spec.template.spec.containers.0.env.0.value: Invalid type. Expected: string, given: integer
The document fixtures/invalid.yaml contains an invalid ReplicationController
--> spec.replicas: Invalid type. Expected: integer, given: string
Missing a kind key in /fixtures/missing-kind.yaml
The document fixtures/valid.json contains a valid Deployment
The document fixtures/valid.yaml contains a valid ReplicationController
```

### From source

If you are modifying `kubeval`, or simply prefer to build your own
binary, then the accompanying `Makefile` has all the build instructions.
If you're on a Mac you should be able to just run:

```
make build
```

The above relies on you having installed Go build environment and
configured `GOPATH`. It also requires `git` to be installed. This will
build binaries in `bin`, and tar files of those binaries in `releases`
for several common architectures.

## Usage

```
$ kubeval --help
Validate a Kubernetes YAML file against the relevant schema

Usage:
  kubeval <file> [file...] [flags]

  Flags:
    -h, --help                        help for kubeval
    -v, --kubernetes-version string   Version of Kubernetes to validate against (default "master")
        --openshift                   Use OpenShift schemas instead of upstream Kubernetes
        --schema-location string      Base URL used to download schemas. Can also be specified with the environment variable KUBEVAL_SCHEMA_LOCATION (default "https://raw.githubusercontent.com/garethr")
        --version                     Display the kubeval version information and exit

```

The command has three important features:

* You can pass one or more files as arguments, including using wildcard
  expansion. Each file will be validated in turn, and `kubeval` will
  exit with a non-zero code if _any_ of the files fail validation.
* You can toggle between the upstream Kubernetes definitions and the
  expanded OpenShift ones using the `--openshift` flag. The default is
  to use the upstream Kubernetes definitions.
* You can pass a version of Kubernetes or OpenShift and the relevant
  type schemas for that version will be used. For instance:

```
$ kubeval -v 1.6.6 my-deployment.yaml
$ kubeval --openshift -v 1.5.1 my-deployment.yaml
```

## Library

After installing with you prefered dependency management tool, import the relevant module.

```go
import (
  "github.com/garethr/kubeval/kubeval"
)
```

The module provides one public function, `Validate`, which can be used
like so:

```go
results, err := kubeval.Validate(fileContents, fileName)
```

The method signature for `Validate` is:

```go
Validate(config []byte, fileName string) ([]ValidationResult, error)
```

The simples way of seeing it's usage is probably in the `kubeval`
[command line tool source code](cmd/root.go).


## Status

`kubeval` should be useful now but can be obviously improved in a number
of ways. If you have suggestions for improvements or new features, or
run into a bug please open issues against the [GitHub
repository](https://github.com/garethr/kubeval). Pull requests also
heartily encouraged.
