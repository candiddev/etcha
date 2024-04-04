// etcha is a CLI tool for managing configurations.
package main

import (
	"os"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/cryptolib"
)

func main() {
	c := config.Default()

	if err := (cli.App[*config.Config]{
		Commands: map[string]cli.Command[*config.Config]{
			"build":    build,
			"compare":  compare,
			"copy":     copyCmd,
			"dir":      dir,
			"file":     file,
			"gen-keys": cryptolib.GenKeys[*config.Config](),
			"jwt":      jwt,
			"init":     initCmd,
			"line":     line,
			"lint":     lint,
			"link":     link,
			"local":    local,
			"push":     push,
			"render":   render,
			"run":      runCmd,
			"shell":    shell,
			"test":     test,
		},
		Config:      c,
		Description: "Etcha is a tool for distributed, serverless build and runtime configuration.",
		Name:        "Etcha",
	}).Run(); err != nil {
		os.Exit(1)
	}
}
