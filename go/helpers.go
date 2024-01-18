package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/etcha/go/metrics"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
)

func dirFileRunParse(args []string) (permissions *fs.FileMode, owner, group *uint32, err error) { //nolint:revive
	if len(args) >= 4 && args[3] != "" {
		p, err := strconv.ParseUint(args[3], 8, 32)
		if err != nil {
			return nil, nil, nil, err
		}

		m := fs.FileMode(p)

		if args[0] == "dir" { //nolint:goconst
			m = fs.FileMode(p + uint64(fs.ModeDir))
		}

		permissions = &m
	}

	if len(args) >= 5 && args[4] != "" {
		o, err := cli.GetUID(args[4])
		if err != nil {
			return nil, nil, nil, err
		}

		owner = &o
	}

	if len(args) >= 6 && args[5] != "" {
		g, err := cli.GetGID(args[5])
		if err != nil {
			return nil, nil, nil, err
		}

		group = &g
	}

	return permissions, owner, group, nil
}

func dirFileRunMk(change, dir bool, contents []byte, path string, permissions *fs.FileMode) (fs.FileInfo, error) {
	f, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) && change { //nolint:revive
			if dir {
				p := fs.FileMode(0755)
				if permissions != nil {
					p = *permissions
				}

				if err := os.Mkdir(path, p); err != nil {
					return nil, err
				}
			} else {
				p := fs.FileMode(0644)
				if permissions != nil {
					p = *permissions
				}

				if err := os.WriteFile(path, contents, p); err != nil {
					return nil, err
				}
			}

			f, err = os.Stat(path)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return f, nil
}

func dirFileRun(file bool, usage string) cli.Command[*config.Config] { //nolint:gocognit,gocyclo
	args := []string{
		"permissions",
		"owner",
		"group",
	}
	if file {
		args = append(args, "contents, or - to read from stdin")
	}

	return cli.Command[*config.Config]{
		ArgumentsRequired: []string{
			"mode [check,change,remove]",
			"path",
		},
		ArgumentsOptional: args,
		Run: func(ctx context.Context, args []string, _ *config.Config) errs.Err {
			dir := args[0] == "dir"
			mode, e := parseMode(args[1], true)
			if e != nil {
				return logger.Error(ctx, e)
			}

			path := args[2]
			permissions, owner, group, err := dirFileRunParse(args)
			if err != nil {
				return logger.Error(ctx, errs.ErrReceiver.Wrap(err))
			}

			var contents []byte

			if !dir && len(args) == 7 {
				if args[6] == "-" {
					contents = cli.ReadStdin()
				} else {
					contents = []byte(args[6])
				}
			}

			// Get file info, or create if it doesn't exist
			f, err := dirFileRunMk(mode == metrics.CommandModeChange, dir, contents, path, permissions)
			if err != nil {
				if mode == metrics.CommandModeRemove {
					return nil
				}

				return logger.Error(ctx, errs.ErrReceiver.Wrap(err))
			}

			if mode == metrics.CommandModeRemove {
				if err := os.Remove(path); err != nil {
					return logger.Error(ctx, errs.ErrReceiver.Wrap(err))
				}

				return nil
			}

			mismatch := []string{}

			if len(contents) > 0 && !dir {
				c, err := os.ReadFile(path)
				if err != nil {
					return logger.Error(ctx, errs.ErrReceiver.Wrap(err))
				}

				if !bytes.Equal(c, contents) {
					if mode == metrics.CommandModeCheck {
						mismatch = append(mismatch, "mismatch contents")
					} else {
						p := fs.FileMode(0644)
						if permissions != nil {
							p = *permissions
						}

						if err := os.WriteFile(path, contents, p); err != nil {
							return logger.Error(ctx, errs.ErrReceiver.Wrap(err))
						}
					}
				}
			}

			if stat, ok := f.Sys().(*syscall.Stat_t); ok {
				g := stat.Gid
				o := stat.Uid

				if group != nil && g != *group {
					mismatch = append(mismatch, fmt.Sprintf("mismatch group: got %d, want %d", g, *group))
					g = *group
				}

				if owner != nil && o != *owner {
					mismatch = append(mismatch, fmt.Sprintf("mismatch owner: got %d, want %d", o, *owner))
					o = *owner
				}

				if (g != stat.Gid || o != stat.Uid) && mode == metrics.CommandModeChange {
					if err := os.Chown(path, int(o), int(g)); err != nil {
						return logger.Error(ctx, errs.ErrReceiver.Wrap(err))
					}
				}
			}

			if permissions != nil && *permissions != f.Mode() {
				mismatch = append(mismatch, fmt.Sprintf("mismatch permissions: got %o, want %o", f.Mode(), *permissions))

				if mode == metrics.CommandModeChange {
					if err := os.Chmod(path, *permissions); err != nil {
						return logger.Error(ctx, errs.ErrReceiver.Wrap(err))
					}
				}
			}

			if len(mismatch) > 0 && mode == metrics.CommandModeCheck {
				return logger.Error(ctx, errs.ErrReceiver.Wrap(fmt.Errorf("file %s does not match:\n\t%s", path, strings.Join(mismatch, "\n\t"))))
			}

			return nil
		},
		Usage: usage,
	}
}

func parseMode(mode string, check bool) (metrics.CommandMode, errs.Err) {
	switch mode {
	case "check":
		if !check {
			break
		}

		return metrics.CommandModeCheck, nil
	case "change":
		return metrics.CommandModeChange, nil
	case "remove":
		return metrics.CommandModeRemove, nil
	}

	return "", errs.ErrReceiver.Wrap(fmt.Errorf("unrecognized mode: %s", mode))
}
