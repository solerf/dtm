package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/solerf/dtm/dotfile"
)

func showResult(action string, status []dotfile.OperationStatus) {
	template := "%s: [%v] to [%v]"

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s summary:\n", action))

	for _, s := range status {
		if s.Error != nil {
			sb.WriteString(fmt.Sprintf("%s: %v\n", fmt.Sprintf(template, s.Status, s.Dotfile.Key, s.Dotfile.SymLink), s.Error))
			continue
		}
		sb.WriteString(fmt.Sprintf(template, s.Status, s.Dotfile.Key, s.Dotfile.SymLink) + "\n")
	}
	log.Println(sb.String())
}

type InstallCmd struct {
	Profiles []string `arg:"" required:"" short:"p" help:"Profiles name to be installed (Profiles name will be matched against dotfiles structure). The order defines the priority"`
	Source   string   `optional:"" short:"s"  type:"path" default:"$PWD" env:"PWD" help:"Path to source directory with dotfiles to be installed."`
	Target   string   `optional:"" short:"t"  type:"path" default:"$HOME" env:"HOME" help:"Path to target directory where dotfiles will be installed."`
}

func (i *InstallCmd) Run() error {
	status, err := dotfile.Install(i.Source, i.Target, i.Profiles...)
	if err != nil {
		return err
	}
	showResult("install", status)
	return nil
}

type UninstallCmd struct {
	Target string `optional:"" short:"t"  type:"path" default:"$HOME" env:"HOME" help:"Path to target directory where dotfiles are installed."`
}

func (u *UninstallCmd) Run() error {
	status, err := dotfile.Uninstall(u.Target)
	if err != nil {
		return err
	}
	showResult("uninstall", status)
	return nil
}

type ShowCmd struct {
	Target string `optional:"" short:"t"  type:"path" default:"$HOME" env:"HOME" help:"Path to target directory where dotfiles are installed."`
}

func (s *ShowCmd) Run() error {
	return dotfile.Show(s.Target)
}

var cmd struct {
	Install   InstallCmd   `cmd:"" help:"Install dotfiles."`
	Uninstall UninstallCmd `cmd:"" help:"Uninstall dotfiles previously installed."`
	Show      ShowCmd      `cmd:"" help:"Show current dotfiles installed."`
}
