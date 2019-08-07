# Developing Kubeval

If you are interested in contributing to Kubeval then the following instructions should
be useful in getting started.

## Pre-requisites

For building Kubeval you'll need the following:

* Make
* Go 1.12

For releasing Kubeval you'll also need:

* [Goreleaser](https://goreleaser.com/)

The acceptance tests use [Bats](https://github.com/sstephenson/bats) and can be run
directly or via Docker.


## Building

Building a binary for your platform can be done by running:

```
make build
```

This should create `bin/kubeval`.


## Testing

The unit tests, along with some basic static analysis, can be run with:

```
make test
```

The [Bats](https://github.com/sstephenson/bats) based acceptance tests
are run using the following target. Note that this runs the tests using Docker.

```
make acceptance
```

If you would prefer to run them directly you need to make sure you have Kubeval
on your PATH and then run:

```
./acceptance.bats
```

