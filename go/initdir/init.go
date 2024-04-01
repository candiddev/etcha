// Package initdir contains functions for initializing an Etcha library.
package initdir

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/jsonnet"
	"github.com/candiddev/shared/go/logger"
)

const readme = `# Etcha Patterns

This repository contains Etcha libraries and patterns written in Jsonnet.  This README and a lot of the files in this repository were initially generated by Etcha.  Please customize to your needs!

## Repository Layout

- ` + "`./lib:`" + ` Jsonnet libraries (files ending with .libsonnet)
- ` + "`./lib/etcha:`" + ` Etcha Jsonnet libraries, probably shouldn't change/edit
- ` + "`./patterns:`" + ` Etcha patterns
`

// Init creates or updates Etcha files in a path.
func Init(ctx context.Context, path string) errs.Err {
	// If path doesn't exist, create it
	if err := os.MkdirAll(path, 0755); err != nil { //nolint:gosec
		return logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("error creating %s", path), err))
	}

	// Populate a README.md if one does not exist
	r := filepath.Join(path, "README.md")
	if _, err := os.Stat(r); errors.Is(err, os.ErrNotExist) {
		if err := os.WriteFile(r, []byte(readme), 0644); err != nil { //nolint:gosec
			return logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("error creating %s", r), err))
		}
	}

	// Create lib directory if not exist
	libd := filepath.Join(path, "lib")
	if err := os.MkdirAll(libd, 0755); err != nil { //nolint:gosec
		return logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("error creating %s", libd), err))
	}

	// Create lib/etcha directory if not exist
	libetcha := filepath.Join(libd, "etcha")
	if err := os.MkdirAll(libetcha, 0755); err != nil { //nolint:gosec
		return logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("error creating %s", libetcha), err))
	}

	// Create pattern directory if not exist
	pattern := filepath.Join(path, "patterns")
	if err := os.MkdirAll(pattern, 0755); err != nil { //nolint:gosec
		return logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("error creating %s", pattern), err))
	}

	// Create native lib
	native := filepath.Join(libetcha, "native.libsonnet")
	if err := os.WriteFile(native, []byte(fmt.Sprintf("// Generated by Etcha %s\n\n", cli.BuildVersion)+jsonnet.Native), 0644); err != nil { //nolint:gosec
		return logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("error creating %s", native), err))
	}

	if err := fs.WalkDir(lib, "lib", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return errors.New("error walking lib: %w")
		}

		if d != nil && !d.Type().IsDir() {
			f, err := lib.ReadFile(path)
			if err != nil {
				return err
			}

			p := filepath.Join(libetcha, strings.Replace(path, "lib/etcha/", "", 1))
			if err := os.WriteFile(p, append([]byte(fmt.Sprintf("// Generated by Etcha %s\n\n", cli.BuildVersion)), f...), 0644); err != nil { //nolint:gosec
				return fmt.Errorf("error creating %s: %w", p, err)
			}
		}

		return nil
	}); err != nil {
		return logger.Error(ctx, errs.ErrReceiver.Wrap(err))
	}

	// Remove unknown lib files
	if err := filepath.Walk(libetcha, func(path string, _ fs.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking path %s: %w", libetcha, err)
		}

		p := filepath.Join("lib/etcha", filepath.Base(path))
		if _, err := lib.Open(p); err != nil && p != "lib/etcha/native.libsonnet" && path != libetcha {
			return os.RemoveAll(path)
		}

		return nil
	}); err != nil {
		return logger.Error(ctx, errs.ErrReceiver.Wrap(err))
	}

	return logger.Error(ctx, nil)
}
