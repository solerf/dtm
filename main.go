package main

import (
	"github.com/alecthomas/kong"
)

var (
	profilePrefix       = "_"
	dtmMappingsFileName = ".dtm_mappings"
)

type Mappings struct {
	SourcesDir []string          `json:"sources_dir"`
	InstallDir string            `json:"install_dir"`
	Entries    map[string]string `json:"entries"`
}

type InstallCmd struct {
	Profile string `arg:"" required:"" short:"p" help:"Profile name to be installed (profile name will be matched against dotfiles structure)."`
	Source  string `optional:"" short:"s"  type:"path" default:"cwd" env:"PWD" help:"Path to source directory with dotfiles to be installed."`
	Target  string `optional:"" short:"t"  type:"path" default:"$HOME" env:"HOME" help:"Path to target directory where dotfiles will be installed."`
}

func (i *InstallCmd) Run() error {
	return Install(i.Profile, i.Source, i.Target)
}

type CleanCmd struct{}

func (d *CleanCmd) Run() error {
	return Clean()
}

var cli struct {
	Install InstallCmd `cmd:"" help:"Install dotfiles at $HOME."`
	Clean   CleanCmd   `cmd:"" help:"Clean dotfiles created at $HOME."`
}

func main() {
	kong.UsageOnError()
	kCtx := kong.Parse(&cli)
	kCtx.FatalIfErrorf(kCtx.Run())
}
