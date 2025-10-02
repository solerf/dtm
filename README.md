dotfiles installer

```shell
Usage: dtm <command>

Flags:
  -h, --help    Show context-sensitive help.

Commands:
  install <profile> [flags]
    Install dotfiles at $HOME.

  clean
    Clean dotfiles created at $HOME.

Run "dtm <command> --help" for more information on a command.
```

```shell
sage: dtm install <profile> [flags]

Install dotfiles at $HOME.

Arguments:
  <profile>    Profile name to be installed (profile name will be matched against
               dotfiles structure).

Flags:
  -h, --help              Show context-sensitive help.

  -s, --source="cwd"      Path to source directory with dotfiles to be installed
                          ($PWD).
  -t, --target="$HOME"    Path to target directory where dotfiles will be installed
                          ($HOME).
```

```shell
Usage: dtm clean

Clean dotfiles created at $HOME.

Flags:
  -h, --help    Show context-sensitive help.
```

// TODO
