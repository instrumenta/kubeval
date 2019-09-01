# Using Kubeval

Kubeval is used to validate one or more Kubernetes configuration files, and
is often used locally as part of a development workflow as well as in CI pipelines.

At the most basic level, Kubeval is used like so:

```console
$ kubeval my-invalid-rc.yaml
WANR - my-invalid-rc.yaml contains an invalid ReplicationController - spec.replicas: Invalid type. Expected: integer, given: string
$ echo $?
1
```

## Strict schemas

The Kubernetes API allows for specifying properties on objects that are not part of the schemas.
However, `kubectl` will throw an error if you use it with such files. Kubeval can be
used to simulate this behaviour using the `--strict` flag.

```console
$ kubeval additional-properties.yaml
PASS - additional-properties.yaml contains a valid ReplicationController
$ echo $?
0
$ kubeval --strict additional-properties.yaml
WARN - additional-properties.yaml contains an invalid ReplicationController - spec: Additional property replicas is not allowed
$ echo $?
1
```

If you're using `kubectl` you may find it useful to always set the `--strict` flag.

## Stdin

Alternatively Kubeval can also take input via `stdin` which can make using
it as part of an automated pipeline easier by removing the need to securely
manage temporary files.

```console
$ cat my-invalid-rc.yaml | kubeval
WARN -  stdin contains an invalid ReplicationController - spec.replicas: Invalid type. Expected: integer, given: string
$ echo $?
1
```

To make the output of pipelines more readable, a filename can be injected
to replace `stdin` in the output:

```console
$ cat my-invalid-rc.yaml | kubeval --filename="my-invalid-rc.yaml"
WARN -  my-invalid-rc.yaml contains an invalid ReplicationController - spec.replicas: Invalid type. Expected: integer, given: string
$ echo $?
1
```

## CRDs

Currently kubeval relies on schemas generated from the Kubernetes API. This means it's not
possible to validate resources using CRDs. Currently you need to pass a flag to ignore
missing schemas, though this may change in a future major version.

```console
$ kubeval --ignore-missing-schemas fixtures/test_crd.yaml
WARN - Set to ignore missing schemas
WARN - fixtures/test_crd.yaml containing a SealedSecret was not validated against a schema
```

If you would prefer to be more explicit about which custom resources to skip you can instead
provide a list of resources to skip like so.

```console
$ kubeval --skip-kinds SealedSecret fixtures/test_crd.yam
WARN - fixtures/test_crd.yaml containing a SealedSecret was not validated against a schema
```

## Helm

Helm chart configurations generally have a reference to the source template in a comment
like so:

```console
# Source: chart/templates/frontend.yam
```

When kubeval detects these comments it will report the relevant chart template files in
the output.

```console
$ kubeval fixtures/multi_valid_source.yaml
PASS - chart/templates/primary.yaml contains a valid Service
PASS - chart/templates/primary.yaml contains a valid ReplicationControlle
```

## Configuring Output

The output of `kubeval` can be configured using the `--output` flag (`-o`).

As of today `kubeval` supports the following output types:

- Plaintext `--output=stdout`
- JSON: `--output=json`
- TAP: `--output=tap`

### Example Output

#### Plaintext

```console
$ kubeval my-invalid-rc.yaml
WARN - my-invalid-rc.yaml contains an invalid ReplicationController - spec.replicas: Invalid type. Expected: integer, given: string
```

#### JSON

```console
$ kubeval fixtures/invalid.yaml -o json
[
     {
             "filename": "fixtures/invalid.yaml",
             "kind": "ReplicationController",
             "status": "invalid",
             "errors": [
                     "spec.replicas: Invalid type. Expected: [integer,null], given: string"
             ]
     }
]
```

#### TAP

```console
 $ kubeval fixtures/invalid.yaml -o tap
1..1
not ok 1 - fixtures/invalid.yaml (ReplicationController) - spec.replicas: Invalid type. Expected: [integer,null], given: string
```

## Full usage instructions

```console
$ kubeval --help
Validate a Kubernetes YAML file against the relevant schema

Usage:
  kubeval <file> [file...] [flags]

Flags:
  -d, --directories strings         A comma-separated list of directories to recursively search for YAML documents
      --exit-on-error               Immediately stop execution when the first error is encountered
  -f, --filename string             filename to be displayed when testing manifests read from stdin (default "stdin")
      --force-color                 Force colored output even if stdout is not a TTY
  -h, --help                        help for kubeval
      --ignore-missing-schemas      Skip validation for resource definitions without a schema
  -v, --kubernetes-version string   Version of Kubernetes to validate against (default "master")
      --openshift                   Use OpenShift schemas instead of upstream Kubernetes
  -o, --output string               The format of the output of this script. Options are: [stdout json]
      --schema-location string      Base URL used to download schemas. Can also be specified with the environment variable KUBEVAL_SCHEMA_LOCATION
      --skip-kinds strings          Comma-separated list of case-sensitive kinds to skip when validating against schemas
      --strict                      Disallow additional properties not in schema
      --version                     version for kubeval
```

The command has three important features:

- You can pass one or more files as arguments, including using wildcard
  expansion. Each file will be validated in turn, and `kubeval` will
  exit with a non-zero code if _any_ of the files fail validation.
- You can toggle between the upstream Kubernetes definitions and the
  expanded OpenShift ones using the `--openshift` flag. The default is
  to use the upstream Kubernetes definitions.
- You can pass a version of Kubernetes or OpenShift and the relevant
  type schemas for that version will be used. For instance:

```console
$ kubeval -v 1.6.6 my-deployment.yaml
$ kubeval --openshift -v 1.5.1 my-deployment.yaml
```
