# Contrib

There are lots of different ways of using Kubeval, this page collects some of those
contributed by users.

## Git pre-commit hook

Add the following to your Kubernetes configs repository in `.git/hooks/pre-commit` to trigger `kubeval` before each commit.

This will validate all the `yaml` files in the top directory of the repository.

```shell
#!/bin/sh

echo "Running kubeval validations..."

if ! [ -x "$(command -v kubeval)" ]; then
  echo 'Error: kubeval is not installed.' >&2
  exit 1
fi

# Inspect code using kubeval
find . -maxdepth 1 -name '*.yaml' -exec kubeval {} \;

status=$?

if [ "$status" = 0 ] ; then
    echo "Static analysis found no problems."
    exit 0
else
    echo 1>&2 "Static analysis found violations that need to be fixed."
    exit 1
fi
```
