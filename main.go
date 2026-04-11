package main

import (
	"github.com/alecthomas/kong"
)

func main() {
	kCtx := kong.Parse(&cmd, kong.UsageOnError())
	kCtx.FatalIfErrorf(kCtx.Run())
}
