# Styling Tests

These tests verify styling and formatting options.

## Setup

```console
$ mkdir -p /tmp/style-test
$ echo "18.17.0" > /tmp/style-test/.nvmrc

```

## Default tmux Output

Default output should include tmux color codes.

```console
$ tmux-package-status -d /tmp/style-test --no-cache | grep -o '#\[fg='
#[fg=
#[fg=

```

## Plain Output (No Color Codes)

Plain output should not include tmux color codes.

```console
$ tmux-package-status -d /tmp/style-test -f plain --no-cache | grep -c '#\[fg=' || echo "0"
0

```

## With Brackets

Test bracket output option.

```console
$ tmux-package-status -d /tmp/style-test -f plain --no-cache --brackets --icons=false
[18.17.0]

```

## Version Prefix

Test version prefix option.

```console
$ tmux-package-status -d /tmp/style-test -f plain --no-cache --icons=false --version-prefix="v"
v18.17.0

```

## Custom Prefix

Test output prefix option.

```console
$ tmux-package-status -d /tmp/style-test -f plain --no-cache --icons=false --prefix=">> "
>> 18.17.0

```

## Custom Suffix

Test output suffix option.

```console
$ tmux-package-status -d /tmp/style-test -f plain --no-cache --icons=false --suffix=" <<"
18.17.0 <<

```

## Custom Separator

Test custom separator with multiple items.

```console
$ echo -e "module test\n\ngo 1.21" > /tmp/style-test/go.mod
$ tmux-package-status -d /tmp/style-test -f plain --no-cache --icons=false --separator=" | " | grep -o ' | '
 |

```

## No Icons

Test disabling icons.

```console
$ rm /tmp/style-test/go.mod
$ tmux-package-status -d /tmp/style-test -f plain --no-cache --icons=false | grep -c '⬢' || echo "0"
0

```

## With Icons

Test that icons are shown by default.

```console
$ tmux-package-status -d /tmp/style-test -f plain --no-cache | grep -o '⬢'
⬢

```

## Environment Variables

Test configuration via environment variables.

```console
$ TMUX_PKG_FORMAT=plain TMUX_PKG_SHOW_ICON=false tmux-package-status -d /tmp/style-test --no-cache
18.17.0

```

## Cleanup

```console
$ rm -rf /tmp/style-test

```
