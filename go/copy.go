package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
	"time"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/etcha/go/metrics"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/cryptolib"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/get"
	"github.com/candiddev/shared/go/logger"
)

var copyCmd = cli.Command[*config.Config]{ //nolint:gochecknoglobals
	ArgumentsRequired: []string{
		"mode [check,change]",
		"src path",
		"dst path or - for stdout",
	},
	Run: func(ctx context.Context, args []string, config *config.Config) errs.Err {
		mode, e := parseMode(args[1], true)
		if e != nil {
			return logger.Error(ctx, e)
		}

		src := args[2]
		dst := args[3]

		var err error

		var checkByte io.ReadWriter

		var dstByte io.ReadWriter

		// Preserve file permissions
		p := fs.FileMode(0644)
		if !strings.HasPrefix(src, "http") {
			f, err := os.Stat(src)
			if err != nil {
				return logger.Error(ctx, errs.ErrReceiver.Wrap(errors.New("error opening src"), err))
			}

			p = f.Mode()
		}

		switch {
		case mode == metrics.CommandModeCheck:
			dstByte = &bytes.Buffer{}

			if dst != "-" {
				checkByte, err = os.Open(dst)
				if err != nil {
					return logger.Error(ctx, errs.ErrReceiver.Wrap(err))
				}
			}
		case dst == "-":
			dstByte = logger.Stdout
		default:
			dstByte, err = os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, p)
			if err != nil {
				return logger.Error(ctx, errs.ErrReceiver.Wrap(err))
			}
		}

		if _, err := get.File(ctx, src, dstByte, time.Time{}); err != nil {
			return logger.Error(ctx, errs.ErrReceiver.Wrap(errors.New("error retreiving src"), err))
		}

		if mode == metrics.CommandModeCheck {
			match := false

			if checkByte != nil {
				s, err := cryptolib.SHA256File(dstByte)
				if err != nil {
					return logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("error checksumming src"), err))
				}

				d, err := cryptolib.SHA256File(checkByte)
				if err != nil {
					return logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("error checksumming dst"), err))
				}

				if s == d {
					match = true
				}
			}

			if !match {
				return logger.Error(ctx, errs.ErrReceiver.Wrap(errors.New("src and dst do not match")))
			}
		}

		return nil
	},
	Usage: "Copy a local file or HTTP path to a destination.  Can optionally specify HTTP headers by appending #<key>:<value>, e.g. #content-type:application/json.",
}
