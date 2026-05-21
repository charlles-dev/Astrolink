package gateway

import (
	"context"
	"time"
)

type Authorization struct {
	MAC        string
	Duration   time.Duration
	DownloadMB int64
	UploadMB   int64
}

type Controller interface {
	Authorize(context.Context, Authorization) error
	Deauthorize(context.Context, string) error
	Ping(context.Context) (time.Duration, error)
}

type CommandRunner interface {
	Run(context.Context, string) (string, error)
}

type NoopController struct{}

func (NoopController) Authorize(context.Context, Authorization) error {
	return nil
}

func (NoopController) Deauthorize(context.Context, string) error {
	return nil
}

func (NoopController) Ping(context.Context) (time.Duration, error) {
	return 0, nil
}
