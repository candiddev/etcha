package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/shared/go/cli"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
	"github.com/itchyny/gojq"
)

var errJQ = errors.New("error querying JSON")

func jq(ctx context.Context, args []string, _ *config.Config) errs.Err {
	s := "."
	raw := false

	for i := range args {
		if i == 0 {
			continue
		}

		if args[i] == "-r" {
			raw = true
		} else {
			s = args[i]
		}
	}

	q, err := gojq.Parse(s)
	if err != nil {
		return logger.Error(ctx, errs.ErrReceiver.Wrap(errJQ, err))
	}

	var v any

	if err := json.Unmarshal([]byte(cli.ReadStdin()), &v); err != nil {
		return logger.Error(ctx, errs.ErrReceiver.Wrap(errJQ, err))
	}

	iter := q.Run(v)

	for {
		v, ok := iter.Next()
		if !ok {
			break
		}

		if err, ok := v.(error); ok {
			return logger.Error(ctx, errs.ErrReceiver.Wrap(errJQ, err))
		}

		m, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return logger.Error(ctx, errs.ErrReceiver.Wrap(errJQ, err))
		}

		s := string(m)

		if raw {
			if strings.HasPrefix(s, `"`) {
				s, err = strconv.Unquote(s)
				if err != nil {
					return logger.Error(ctx, errs.ErrReceiver.Wrap(errJQ, err))
				}
			}
		}

		logger.Raw(fmt.Sprintf("%s\n", s))
	}

	return nil
}
