// etcha is a CLI tool for managing configurations.
package main

import (
	"os"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/crypto"
)

func main() {
	c := config.Default()

	if err := (cli.App[*config.Config]{
		Commands: map[string]cli.Command[*config.Config]{
			"build": {
				ArgumentsRequired: []string{
					"pattern path",
					"destination jwt",
				},
				ArgumentsOptional: []string{
					"config source, default: etcha",
				},
				Run:   build,
				Usage: "Build a pattern",
			},
			"check": {
				ArgumentsRequired: []string{
					"pattern path",
				},
				ArgumentsOptional: []string{
					"config source, default: etcha",
				},
				Run:   runCommands,
				Usage: "Check run commands in a pattern as a one-off apply",
			},
			"compare": {
				ArgumentsRequired: []string{
					"new jwt path or URL",
					"old jwt path or URL",
				},
				ArgumentsOptional: []string{
					"ignore version, default: no",
				},
				Run:   compare,
				Usage: "Compare two JWTs to see if they have the same etchaBuildManifest, etchaPattern, and optionally etchaVersion",
			},
			"generate-ed25519": {
				Run:   crypto.GenerateEd25519[*config.Config],
				Usage: "Generate an Ed25519 keypair",
			},
			"init": {
				ArgumentsOptional: []string{
					"directory, default: current directory",
				},
				Run:   initDir,
				Usage: "Initialize a directory for pattern development",
			},
			"lint": {
				ArgumentsRequired: []string{
					"pattern path or directory",
				},
				ArgumentsOptional: []string{
					"check formatting, default: no",
				},
				Run:   lint,
				Usage: "Lint a pattern or directory",
			},
			"push": {
				ArgumentsRequired: []string{
					"pattern path",
					"destination URL",
				},
				Run:   push,
				Usage: "Push a signed pattern JWT to a destination URL",
			},
			"remove": {
				ArgumentsRequired: []string{
					"pattern path",
				},
				ArgumentsOptional: []string{
					"config source, default: etcha",
				},
				Run:   runCommands,
				Usage: "Remove run commands in a pattern as a one-off apply",
			},
			"run-commands": {
				ArgumentsRequired: []string{
					"pattern path",
				},
				ArgumentsOptional: []string{
					"config source, default: etcha",
				},
				Run:   runCommands,
				Usage: "Run commands in a pattern as a one-off apply",
			},
			"run-listen": {
				Run:   runCommands,
				Usage: "Run Etcha in listening mode, periodically pulling new patterns or receiving new patterns via push",
			},
			"run-once": {
				Run:   runCommands,
				Usage: "Run Etcha patterns once, pull new patterns, and run the diff",
			},
			"show-pattern": {
				ArgumentsRequired: []string{
					"jwt or pattern path",
				},
				Run:   showCommands,
				Usage: "Show the rendered pattern of a JWT or pattern file",
			},
			"show-jwt": {
				ArgumentsRequired: []string{
					"jwt path",
				},
				Run:   showJWT,
				Usage: "Show the contents of a JWT",
			},
			"test": {
				ArgumentsRequired: []string{
					"pattern path or directory",
				},
				Run:   test,
				Usage: "Test a pattern or directory",
			},
		},
		Config:      c,
		Description: "Etcha is a tool for distributed, serverless build and runtime configuration.",
		Name:        "Etcha",
	}).Run(); err != nil {
		os.Exit(1)
	}
}
