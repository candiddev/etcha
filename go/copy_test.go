package main

import (
	"context"
	"os"
	"testing"

	"github.com/candiddev/etcha/go/config"
	"github.com/candiddev/shared/go/assert"
	"github.com/candiddev/shared/go/errs"
	"github.com/candiddev/shared/go/logger"
)

func TestCopyCmd(t *testing.T) {
	ctx := context.Background()
	ctx = logger.SetNoColor(ctx, true)
	c := config.Default()

	os.WriteFile("test", []byte("hello"), 0600)

	logger.SetStd()
	assert.HasErr(t, copyCmd.Run(ctx, []string{
		"copy",
		"check",
		"test",
		"test2",
	}, nil, c), errs.ErrReceiver)
	assert.Equal(t, logger.ReadStd(), "ERROR open test2: no such file or directory\n")

	logger.SetStd()
	assert.HasErr(t, copyCmd.Run(ctx, []string{
		"copy",
		"change",
		"test",
		"-",
	}, nil, c), nil)
	assert.Equal(t, logger.ReadStd(), "hello")

	logger.SetStd()
	assert.HasErr(t, copyCmd.Run(ctx, []string{
		"copy",
		"change",
		"test",
		"test2",
	}, nil, c), nil)
	assert.Equal(t, logger.ReadStd(), "")

	logger.SetStd()
	assert.HasErr(t, copyCmd.Run(ctx, []string{
		"copy",
		"check",
		"test",
		"test2",
	}, nil, c), nil)
	assert.Equal(t, logger.ReadStd(), "")

	logger.SetStd()
	assert.HasErr(t, copyCmd.Run(ctx, []string{
		"copy",
		"check",
		"test",
		"-",
	}, nil, c), errs.ErrReceiver)
	assert.Equal(t, logger.ReadStd(), "ERROR src and dst do not match\n")

	logger.SetStd()
	assert.HasErr(t, copyCmd.Run(ctx, []string{
		"copy",
		"change",
		"/testttttt",
		"-",
	}, nil, c), errs.ErrReceiver)
	assert.Equal(t, logger.ReadStd(), "ERROR error opening src: stat /testttttt: no such file or directory\n")

	os.Remove("test")
	os.Remove("test2")
}
