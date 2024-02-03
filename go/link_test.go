package main

import (
	"context"
	"os"
	"testing"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/shared/go/assert"
	"github.com/candiddev/shared/go/logger"
)

func TestLink(t *testing.T) {
	logger.UseTestLogger(t)

	ctx := context.Background()
	ctx = logger.SetFormat(ctx, logger.FormatKV)
	c := config.Default()

	os.WriteFile("test", []byte{}, 0700)

	tests := []struct {
		name    string
		mode    string
		src     string
		dst     string
		wantErr bool
	}{
		{
			name:    "check_missing",
			mode:    "check",
			src:     "/test1",
			dst:     "test",
			wantErr: true,
		},
		{
			name:    "change_error1",
			mode:    "change",
			src:     "/test1",
			dst:     "/test",
			wantErr: true,
		},
		{
			name:    "change_error2",
			mode:    "change",
			src:     "/test1",
			dst:     "/root",
			wantErr: true,
		},
		{
			name: "change_ok1",
			mode: "change",
			src:  "/test1",
			dst:  "test",
		},
		{
			name: "change_ok2",
			mode: "change",
			src:  "/test2",
			dst:  "test",
		},
		{
			name: "change_ok3",
			mode: "change",
			src:  "/test2",
			dst:  "test",
		},
		{
			name: "check_ok",
			mode: "check",
			src:  "/test2",
			dst:  "test",
		},
		{
			name: "remove_ok",
			mode: "remove",
			src:  "/test2",
			dst:  "test",
		},
		{
			name: "remove_ok",
			mode: "remove",
			src:  "/test2",
			dst:  "test",
		},
		{
			name:    "remove_error",
			mode:    "remove",
			src:     "/test1",
			dst:     "/bin",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := link.Run(ctx, []string{
				"",
				tc.mode,
				tc.src,
				tc.dst,
			}, c)
			assert.Equal(t, err != nil, tc.wantErr)
		})
	}
}
