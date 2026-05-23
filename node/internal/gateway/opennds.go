package gateway

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	ErrInvalidMAC      = errors.New("mac invalido")
	ErrInvalidDuration = errors.New("duracao invalida")
	macPattern         = regexp.MustCompile(`^[0-9A-F]{2}(:[0-9A-F]{2}){5}$`)
)

type OpenNDSOptions struct {
	Retries    int
	RetryDelay time.Duration
}

type OpenNDSController struct {
	runner CommandRunner
	opts   OpenNDSOptions
}

func NewOpenNDSController(runner CommandRunner, opts OpenNDSOptions) *OpenNDSController {
	if opts.Retries <= 0 {
		opts.Retries = 1
	}
	if opts.RetryDelay <= 0 {
		opts.RetryDelay = 200 * time.Millisecond
	}
	return &OpenNDSController{runner: runner, opts: opts}
}

func (c *OpenNDSController) Authorize(ctx context.Context, input Authorization) error {
	mac, err := normalizeMAC(input.MAC)
	if err != nil {
		return err
	}
	if input.Duration <= 0 {
		return ErrInvalidDuration
	}
	seconds := int64(input.Duration.Round(time.Second).Seconds())
	command := fmt.Sprintf(
		"ndsctl auth %s %d %d %d",
		mac,
		seconds,
		megabytesToBytes(input.DownloadMB),
		megabytesToBytes(input.UploadMB),
	)
	return c.runWithRetry(ctx, command)
}

func (c *OpenNDSController) Deauthorize(ctx context.Context, mac string) error {
	normalized, err := normalizeMAC(mac)
	if err != nil {
		return err
	}
	return c.runWithRetry(ctx, fmt.Sprintf("ndsctl deauth %s", normalized))
}

func (c *OpenNDSController) Ping(ctx context.Context) (time.Duration, error) {
	start := time.Now()
	err := c.runWithRetry(ctx, "ndsctl status")
	return time.Since(start), err
}

func (c *OpenNDSController) Diagnostic(ctx context.Context) (RouterDiagnostic, error) {
	statusOutput, err := c.runOutputWithRetry(ctx, "ndsctl status")
	if err != nil {
		return RouterDiagnostic{}, err
	}
	clientsOutput, err := c.runOutputWithRetry(ctx, "ndsctl clients")
	if err != nil {
		return RouterDiagnostic{}, err
	}
	boardOutput, err := c.runOutputWithRetry(ctx, "ubus call system board")
	if err != nil {
		return RouterDiagnostic{}, err
	}
	logsOutput, err := c.runOutputWithRetry(ctx, "logread -e opennds -n 50")
	if err != nil {
		return RouterDiagnostic{}, err
	}
	return BuildRouterDiagnostic(statusOutput, clientsOutput, boardOutput, logsOutput), nil
}

func (c *OpenNDSController) runWithRetry(ctx context.Context, command string) error {
	_, err := c.runOutputWithRetry(ctx, command)
	return err
}

func (c *OpenNDSController) runOutputWithRetry(ctx context.Context, command string) (string, error) {
	var lastErr error
	for attempt := 0; attempt < c.opts.Retries; attempt++ {
		output, err := c.runner.Run(ctx, command)
		if err == nil {
			return output, nil
		}
		lastErr = err
		if attempt == c.opts.Retries-1 {
			break
		}
		timer := time.NewTimer(c.opts.RetryDelay * time.Duration(attempt+1))
		select {
		case <-ctx.Done():
			timer.Stop()
			return "", ctx.Err()
		case <-timer.C:
		}
	}
	return "", fmt.Errorf("executar %q: %w", command, lastErr)
}

func normalizeMAC(mac string) (string, error) {
	normalized := strings.ToUpper(strings.TrimSpace(mac))
	normalized = strings.ReplaceAll(normalized, "-", ":")
	if !macPattern.MatchString(normalized) {
		return "", ErrInvalidMAC
	}
	return normalized, nil
}

func megabytesToBytes(value int64) int64 {
	if value <= 0 {
		return 0
	}
	return value * 1024 * 1024
}
