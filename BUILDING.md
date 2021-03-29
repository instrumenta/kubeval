# Developing Kubeval

If you are interested in contributing to Kubeval then the following instructions should
be useful in getting started.

## Pre-requisites

For building Kubeval you'll need the following:

* Make
* Go 1.15

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

### Release Snapshot

To build the release snapshots run:

```
make snapshot
```

This creates the directory `dist` with all available release artifacts and the final configuration for `goreleaser`.

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

If you would prefer to run them directly

```
make build && PATH=./bin:$PATH ./acceptance.bats
```

