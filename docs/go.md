# Go Library

Kubeval is implemented in Go, and can be used as a Go library as well as being
used as a command line tool.

The module can be imported like so:


```go
import (
  "github.com/instrumenta/kubeval/kubeval"
)
```

The module provides one public function, `Validate`, which can be used
like so:

```go
results, err := kubeval.Validate(fileContents, fileName)
```

The method signature for `Validate` is:

```go
Validate(input []byte, config kubeval.Config) ([]ValidationResult, error)
```

The simplest way of seeing it's usage is probably in the `kubeval`
[command line tool source code](https://github.com/instrumenta/kubeval/blob/master/main.go).
