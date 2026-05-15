# DTM

Simple dotfiles installer, allows to merge and install different dotfiles profiles.

```shell
Usage: dtm <command>

Flags:
  -h, --help    Show context-sensitive help.

Commands:
  install <profiles> ... [flags]
    Install dotfiles.

  uninstall [flags]
    Uninstall dotfiles previously installed.

  show [flags]
    Show current dotfiles installed.

Run "dtm <command> --help" for more information on a command.
```

```shell
Usage: dtm install <profiles> ... [flags]

Install dotfiles.

Arguments:
  <profiles> ...    Profiles name to be installed (Profiles name will be matched
                    against dotfiles structure). The order defines the priority

Flags:
  -h, --help              Show context-sensitive help.

  -s, --source="$PWD"     Path to source directory with dotfiles to be installed
                          ($PWD).
  -t, --target="$HOME"    Path to target directory where dotfiles will be
                          installed ($HOME).
```

```shell
Usage: dtm uninstall [flags]

Uninstall dotfiles previously installed.

Flags:
  -h, --help              Show context-sensitive help.

  -t, --target="$HOME"    Path to target directory where dotfiles are installed
                          ($HOME).
```

```shell
Usage: dtm show [flags]

Show current dotfiles installed.

Flags:
  -h, --help              Show context-sensitive help.

  -t, --target="$HOME"    Path to target directory where dotfiles are installed
                          ($HOME).
```
