package run

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/etcha/go/metrics"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/jwt"
	"github.com/candiddev/shared/go/logger"
	"github.com/candiddev/shared/go/types"
	"github.com/creack/pty"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang.org/x/term"
)

type shellSession struct {
	ID    string `json:"i,omitempty"`
	Stdin string `json:"s,omitempty"`
}

func Shell(ctx context.Context, c *config.Config, target config.Target, source string) errs.Err {
	d := (&url.URL{
		Host: net.JoinHostPort(target.Hostname, strconv.Itoa(target.Port)),
		Path: target.PathShell + "/" + source,
	})

	if target.Insecure {
		d.Scheme = "http"
	} else {
		d.Scheme = "https"
	}

	u := d.String()

	req, er := postShell(ctx, c, u, "", nil)
	if er != nil {
		return logger.Error(ctx, er)
	}

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Connection", "keep-alive")

	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: c.Build.TLSSkipVerify, //nolint:gosec
			},
		},
	}

	res, err := client.Do(req)
	if err != nil {
		return logger.Error(ctx, errs.ErrReceiver.Wrap(errors.New("error performing request"), err))
	}

	defer res.Body.Close()

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return logger.Error(ctx, errs.ErrReceiver.Wrap(errors.New("error reseting terminal"), err))
	}

	defer term.Restore(int(os.Stdin.Fd()), oldState) //nolint:errcheck

	term.NewTerminal(struct {
		io.Reader
		io.Writer
	}{
		os.Stdin,
		os.Stdout,
	}, "")

	b := bufio.NewReader(res.Body)

	out, err := b.ReadString('\n')
	if err != nil {
		return logger.Error(ctx, errs.ErrReceiver.Wrap(errors.New("error reading id"), err))
	}

	id := out[:len(out)-1]

	go func() {
		b := make([]byte, 4096)

		for {
			n, er := os.Stdin.Read(b)
			if er != nil && er != io.EOF {
				return
			}

			req, err := postShell(ctx, c, u, id, b[:n])
			if err == nil {
				res, er := client.Do(req)
				if er != nil {
					logger.Error(ctx, errs.ErrReceiver.Wrap(errors.New("error performing request"), err)) //nolint:errcheck
				}

				res.Body.Close()
			} else {
				logger.Error(ctx, err) //nolint:errcheck
			}
		}
	}()

	for {
		out, err := b.ReadString('\n')
		if err != nil {
			return logger.Error(ctx, errs.ErrReceiver.Wrap(err))
		}

		if len(out) > 1 && string(out[0]) == "o" {
			out, err := base64.StdEncoding.DecodeString(out[1:])
			if err == nil {
				fmt.Fprint(os.Stdout, string(out)) //nolint:forbidigo
			}
		}
	}
}

func postShell(ctx context.Context, c *config.Config, u, id string, stdin []byte) (*http.Request, errs.Err) {
	t, _, err := jwt.New(shellSession{
		ID:    id,
		Stdin: base64.StdEncoding.EncodeToString(stdin),
	}, time.Now().Add(5*time.Second), nil, "", "", "")
	if err != nil {
		return nil, logger.Error(ctx, errs.ErrReceiver.Wrap(err))
	}

	b, er := c.SignJWT(ctx, t)
	if er != nil {
		return nil, logger.Error(ctx, er)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewBufferString(b))
	if err != nil {
		return nil, logger.Error(ctx, errs.ErrReceiver.Wrap(errors.New("error creating request"), err))
	}

	return req, nil
}

func (s *state) postShell(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	srcName := chi.URLParam(r, "source")
	ctx = metrics.SetSourceName(ctx, srcName)
	ctx = logger.SetAttribute(ctx, "remoteAddr", r.RemoteAddr)

	src, ok := s.Config.Sources[srcName]
	if !ok || src.Shell == "" {
		logger.Error(ctx, errs.ErrSenderBadRequest.Wrap(fmt.Errorf("unknown source: %s", srcName))) //nolint: errcheck
		w.WriteHeader(http.StatusNotFound)

		return
	}

	body, e := io.ReadAll(r.Body)
	if e != nil {
		logger.Error(ctx, errs.ErrSenderBadRequest.Wrap(errors.New("error reading body"), e)) //nolint: errcheck
		w.WriteHeader(http.StatusNotFound)

		return
	}

	sh := shellSession{}

	k, _, err := s.Config.ParseJWT(ctx, &sh, string(body), srcName)
	if err != nil {
		logger.Error(ctx, errs.ErrReceiver.Wrap(err)) //nolint: errcheck
		w.WriteHeader(http.StatusNotFound)

		return
	}

	if p := s.Shells.Get(sh.ID); p != nil {
		out, err := base64.StdEncoding.DecodeString(sh.Stdin)
		if err != nil {
			logger.Error(ctx, errs.ErrReceiver.Wrap(errors.New("error decoding stdin"), err)) //nolint:errcheck
		} else {
			if _, err := p.Write(out); err != nil {
				logger.Error(ctx, errs.ErrReceiver.Wrap(errors.New("error writing shell to stdin"), err)) //nolint:errcheck
			}
		}

		return
	}

	logger.Info(ctx, fmt.Sprintf("Shell session started for source %s using key %s", srcName, k.ID))

	// Check if SSE is supported
	flusher, ok := w.(http.Flusher)
	if !ok {
		logger.Error(ctx, errs.ErrReceiver.Wrap(errors.New("error creating new pipe"), err)) //nolint: errcheck
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	id := uuid.New()

	fmt.Fprintf(w, "%s\n", id.String()) //nolint:forbidigo
	flusher.Flush()

	// Create Pty
	exec := s.Config.Exec.Override(src.Exec)
	exec.Command = src.Shell

	opts, err := exec.RunOpts(ctx, "")
	if err != nil {
		logger.Error(ctx, errs.ErrReceiver.Wrap(errors.New("error getting shell configuration"), err)) //nolint: errcheck
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	cmd, err := opts.GetCmd(ctx)
	if err != nil {
		logger.Error(ctx, errs.ErrReceiver.Wrap(errors.New("error getting shell command"), err)) //nolint: errcheck
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	p, err := pty.Start(cmd)
	if err != nil {
		logger.Error(ctx, errs.ErrReceiver.Wrap(errors.New("error starting shell"), err)) //nolint: errcheck
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	defer p.Close()

	go func() {
		b := make([]byte, 4096)

		for {
			n, err := p.Read(b)
			if err != nil && err != io.EOF {
				return
			}

			fmt.Fprintf(w, "o%s\n", base64.StdEncoding.EncodeToString(b[:n])) //nolint:forbidigo
			flusher.Flush()
		}
	}()

	s.Shells.Set(id.String(), p)
	defer s.Shells.Set(id.String(), nil)

	done := make(chan bool)

	go func() {
		_, err := exec.Run(ctx, s.Config.CLI, "")
		if err != nil {
			fmt.Fprintf(w, "o%s\n", err.Error()) //nolint:forbidigo
		}

		done <- true
	}()

	tick := time.NewTicker(time.Second * time.Duration((types.RandInt(10) + 1)))

	for {
		select {
		case <-done:
			return
		case <-tick.C:
			tick.Reset(time.Second * time.Duration((types.RandInt(10) + 1)))
			fmt.Fprintf(w, "k%s\n", types.RandString(int(types.RandInt(20)))) //nolint:forbidigo
			flusher.Flush()
		}
	}
}
