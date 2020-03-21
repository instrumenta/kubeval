# Installing Kubeval

Tagged versions of `kubeval` are built using [GoReleaser](https://goreleaser.com/) and
uploaded to GitHub. This means you should find `tar.gz` and `.zip` files under the
release tab. These should contain a single `kubeval` binary for the platform
in the filename (ie. windows, linux, darwin). Either execute that binary
directly or place it on your path.


## Linux

```
wget https://github.com/instrumenta/kubeval/releases/latest/download/kubeval-linux-amd64.tar.gz
tar xf kubeval-linux-amd64.tar.gz
sudo cp kubeval /usr/local/bin
```

## macOS

```
wget https://github.com/instrumenta/kubeval/releases/latest/download/kubeval-darwin-amd64.tar.gz
tar xf kubeval-darwin-amd64.tar.gz
sudo cp kubeval /usr/local/bin
```

For those using [Homebrew](https://brew.sh/) you can use the kubeval tap:

```
brew tap instrumenta/instrumenta
brew install kubeval
```

## Windows

Windows users can download the `zip` files from the releases page. For [Scoop](https://scoop.sh/)
users you can install with:

```
scoop bucket add instrumenta https://github.com/instrumenta/scoop-instrumenta
scoop install kubeval
```

## Docker


`kubeval` is also published as a Docker image. This can be used as follows:
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
