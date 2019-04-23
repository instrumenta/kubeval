# Using Kubeval

Kubeval is used to validate one or more Kubernetes configuration files, and
is often used locally as part of a development workflow as well as in CI pipelines.

At the most basic level, Kubeval is used like so:

```console
$ kubeval my-invalid-rc.yaml
The document my-invalid-rc.yaml contains an invalid ReplicationController
--> spec.replicas: Invalid type. Expected: integer, given: string
$ echo $?
1
```


## Strict schemas

The Kubernetes API allows for specifying properties on objects that are not part of the schemas.
However, `kubectl` will throw an error if you use it with such files. Kubeval can be
used to simulate this behaviour using the `--strict` flag.

```console
$ kubeval additional-properties.yaml
The document additional-properties.yaml contains a valid ReplicationController
$ echo $?
0
$ kubeval --strict additional-properties.yaml
The document additional-properties.yaml contains an invalid ReplicationController
---> spec: Additional property replicas is not allowed
$ echo $?
1
```

If you're using `kubectl` you may find it useful to always set the `--strict` flag.


## Stdin

Alternatively Kubeval can also take input via `stdin` which can make using
it as part of an automated pipeline easier by removing the need to securely
manage temporary files.

```
$ cat my-invalid-rc.yaml | kubeval
The document stdin contains an invalid ReplicationController
--> spec.replicas: Invalid type. Expected: integer, given: string
$ echo $?
1
```

To make the output of pipelines more readable, a filename can be injected
to replace `stdin` in the output:

```
$ cat my-invalid-rc.yaml | kubeval --filename="my-invalid-rc.yaml"
The document my-invalid-rc.yaml contains an invalid ReplicationController
--> spec.replicas: Invalid type. Expected: integer, given: string
$ echo $?
1
```


## Full usage instructions

```
$ kubeval --help
Validate a Kubernetes YAML file against the relevant schema

Usage:
  kubeval <file> [file...] [flags]

Flags:
  -f, --filename string             filename to be displayed when testing manifests read from stdin (default "stdin")
  -h, --help                        help for kubeval
  -v, --kubernetes-version string   Version of Kubernetes to validate against (default "master")
      --openshift                   Use OpenShift schemas instead of upstream Kubernetes
      --schema-location string      Base URL used to download schemas. Can also be specified with the environment variable KUBEVAL_SCHEMA_LOCATION (default "https://raw.githubusercontent.com/garethr")
      --strict                      Disallow additional properties not in schema
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
