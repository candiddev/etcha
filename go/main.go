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
			"build": {
				ArgumentsRequired: []string{
					"pattern path",
					"destination path",
				},
				ArgumentsOptional: []string{
					"config source, default: etcha",
				},
				Run:   build,
				Usage: "Build a pattern",
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
			"generate-keys": cryptolib.GenerateKeys[*config.Config](),
			"init": {
				ArgumentsOptional: []string{
					"directory, default: current directory",
				},
				Run:   initDir,
				Usage: "Initialize a directory for pattern development",
			},
			"jq": {
				ArgumentsOptional: []string{
					"jq query options",
				},
				Run:   jq,
				Usage: "Query JSON from stdin using jq.  Supports standard JQ queries, and the -r flag to render raw values.",
			},
			"lint": {
				ArgumentsRequired: []string{
					"path",
				},
				ArgumentsOptional: []string{
					"check formatting, default: no",
				},
				Run:   lint,
				Usage: "Lint a pattern or directory",
			},
			"local-change": {
				ArgumentsRequired: []string{
					"pattern path",
				},
				ArgumentsOptional: []string{
					"config source, default: etcha",
				},
				Run:   runCommands,
				Usage: "Run commands in a pattern locally",
			},
			"local-remove": {
				ArgumentsRequired: []string{
					"pattern path",
				},
				ArgumentsOptional: []string{
					"config source, default: etcha",
				},
				Run:   runCommands,
				Usage: "Run remove commands in a pattern locally",
			},
			"push-command": {
				ArgumentsRequired: []string{
					"command",
					"destination URL",
				},
				Run:   push,
				Usage: "Push a signed command to a destination URL",
			},
			"push-pattern": {
				ArgumentsRequired: []string{
					"pattern path",
					"destination URL",
				},
				Run:   push,
				Usage: "Push a signed pattern JWT to a destination URL",
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
					"path",
				},
				ArgumentsOptional: []string{
					"test build commands, default: no",
				},
				Run:   test,
				Usage: "Test all patterns in path",
			},
		},
		Config:      c,
		Description: "Etcha is a tool for distributed, serverless build and runtime configuration.",
		Name:        "Etcha",
	}).Run(); err != nil {
		os.Exit(1)
	}
}
