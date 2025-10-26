package main

import (
	"github.com/solerf/dtm/runner"
)

type InstallCmd struct {
	Profiles []string `arg:"" required:"" short:"p" help:"Profiles name to be installed (Profiles name will be matched against dotfiles structure). The order defines the priority"`
	Source   string   `optional:"" short:"s"  type:"path" default:"cwd" env:"PWD" help:"Path to source directory with dotfiles to be installed."`
	Target   string   `optional:"" short:"t"  type:"path" default:"$HOME" env:"HOME" help:"Path to target directory where dotfiles will be installed."`
}

func (i *InstallCmd) Run() error {
	if err := runner.Install(i.Profiles, i.Source, i.Target); err != nil {
		return err
	}
	return nil
}

type UninstallCmd struct{}

func (d *UninstallCmd) Run() error {
	if err := runner.Uninstall(); err != nil {
		return err
	}
	return nil
}

var cmd struct {
	Install   InstallCmd   `cmd:"" help:"Install dotfiles at $HOME."`
	Uninstall UninstallCmd `cmd:"" help:"Uninstall dotfiles created at $HOME."`
}
