# Detection Tests

These tests verify package file detection and version parsing.

## Setup Test Directory

Create a test directory with various package files.

```console
$ mkdir -p /tmp/pkg-test && cd /tmp/pkg-test

```

## Node.js Detection (nvmrc)

Test detection of Node.js version from .nvmrc file.

```console
$ echo "18.17.0" > /tmp/pkg-test/.nvmrc
$ tmux-package-status -d /tmp/pkg-test -f plain --no-cache --icons=false
18.17.0

```

## Node.js Detection with v prefix

Test that v prefix is stripped from version.

```console
$ echo "v20.10.0" > /tmp/pkg-test/.nvmrc
$ tmux-package-status -d /tmp/pkg-test -f plain --no-cache --icons=false
20.10.0

```

## JSON Output

Test JSON output format.

```console
$ echo "18.17.0" > /tmp/pkg-test/.nvmrc
$ tmux-package-status -d /tmp/pkg-test -f json --no-cache | grep -o '"type":"node"'
"type":"node"

```

```console
$ tmux-package-status -d /tmp/pkg-test -f json --no-cache | grep -o '"version":"18.17.0"'
"version":"18.17.0"

```

## Go Detection

Test detection of Go version from go.mod file.

```console
$ rm -f /tmp/pkg-test/.nvmrc
$ echo -e "module test\n\ngo 1.21" > /tmp/pkg-test/go.mod
$ tmux-package-status -d /tmp/pkg-test -f plain --no-cache --icons=false
1.21

```

## Go Detection with Patch Version

Test detection of Go version with patch number.

```console
$ echo -e "module test\n\ngo 1.21.5" > /tmp/pkg-test/go.mod
$ tmux-package-status -d /tmp/pkg-test -f plain --no-cache --icons=false
1.21.5

```

## Python Detection

Test detection of Python version from .python-version file.

```console
$ rm -f /tmp/pkg-test/go.mod
$ echo "3.11.5" > /tmp/pkg-test/.python-version
$ tmux-package-status -d /tmp/pkg-test -f plain --no-cache --icons=false
3.11.5

```

## Ruby Detection

Test detection of Ruby version from .ruby-version file.

```console
$ rm -f /tmp/pkg-test/.python-version
$ echo "3.2.2" > /tmp/pkg-test/.ruby-version
$ tmux-package-status -d /tmp/pkg-test -f plain --no-cache --icons=false
3.2.2

```

## Ruby Detection with ruby- prefix

Test that ruby- prefix is stripped.

```console
$ echo "ruby-3.2.2" > /tmp/pkg-test/.ruby-version
$ tmux-package-status -d /tmp/pkg-test -f plain --no-cache --icons=false
3.2.2

```

## Multiple Package Files

Test detection of multiple package managers.

```console
$ rm -f /tmp/pkg-test/.ruby-version
$ echo "18.17.0" > /tmp/pkg-test/.nvmrc
$ echo -e "module test\n\ngo 1.21" > /tmp/pkg-test/go.mod
$ tmux-package-status -d /tmp/pkg-test -f json --no-cache | grep -c '"type"'
2

```

## Filter by Manager

Test filtering to show only specific managers.

```console
$ tmux-package-status -d /tmp/pkg-test -f plain --no-cache --icons=false -m node
18.17.0

```

```console
$ tmux-package-status -d /tmp/pkg-test -f plain --no-cache --icons=false -m go
1.21

```

## Max Items

Test limiting the number of items shown.

```console
$ echo "3.11.5" > /tmp/pkg-test/.python-version
$ tmux-package-status -d /tmp/pkg-test -f json --no-cache --max 2 | grep -c '"type"'
2

```

## Cleanup

```console
$ rm -rf /tmp/pkg-test

```
