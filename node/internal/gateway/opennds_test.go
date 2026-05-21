package gateway_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/astrolink/node/internal/gateway"
)

func TestOpenNDSController_Authorize_FormatsNDSAuthCommand(t *testing.T) {
	runner := &recordingRunner{}
	controller := gateway.NewOpenNDSController(runner, gateway.OpenNDSOptions{Retries: 1})

	err := controller.Authorize(context.Background(), gateway.Authorization{
		MAC:        "aa:bb:cc:dd:ee:ff",
		Duration:   24 * time.Hour,
		DownloadMB: 100,
		UploadMB:   50,
	})

	if err != nil {
		t.Fatal(err)
	}
	want := "ndsctl auth AA:BB:CC:DD:EE:FF 86400 104857600 52428800"
	if runner.commands[0] != want {
		t.Fatalf("command = %q, want %q", runner.commands[0], want)
	}
}

func TestOpenNDSController_Deauthorize_FormatsNDSDeauthCommand(t *testing.T) {
	runner := &recordingRunner{}
	controller := gateway.NewOpenNDSController(runner, gateway.OpenNDSOptions{Retries: 1})

	err := controller.Deauthorize(context.Background(), "aa:bb:cc:dd:ee:ff")

	if err != nil {
		t.Fatal(err)
	}
	want := "ndsctl deauth AA:BB:CC:DD:EE:FF"
	if runner.commands[0] != want {
		t.Fatalf("command = %q, want %q", runner.commands[0], want)
	}
}

func TestOpenNDSController_Authorize_RejectsInvalidMAC(t *testing.T) {
	runner := &recordingRunner{}
	controller := gateway.NewOpenNDSController(runner, gateway.OpenNDSOptions{Retries: 1})

	err := controller.Authorize(context.Background(), gateway.Authorization{
		MAC:      "not-a-mac",
		Duration: time.Hour,
	})

	if err == nil {
		t.Fatal("expected invalid MAC error")
	}
	if len(runner.commands) != 0 {
		t.Fatalf("expected no commands, got %v", runner.commands)
	}
}

func TestOpenNDSController_Authorize_RetriesTransientCommandFailure(t *testing.T) {
	runner := &recordingRunner{failures: 2}
	controller := gateway.NewOpenNDSController(runner, gateway.OpenNDSOptions{
		Retries:    3,
		RetryDelay: time.Nanosecond,
	})

	err := controller.Authorize(context.Background(), gateway.Authorization{
		MAC:      "AA:BB:CC:DD:EE:FF",
		Duration: time.Hour,
	})

	if err != nil {
		t.Fatal(err)
	}
	if runner.calls != 3 {
		t.Fatalf("calls = %d, want 3", runner.calls)
	}
}

type recordingRunner struct {
	commands []string
	failures int
	calls    int
}

func (r *recordingRunner) Run(_ context.Context, command string) (string, error) {
	r.calls++
	r.commands = append(r.commands, command)
	if r.failures > 0 {
		r.failures--
		return "", errors.New("temporary ssh failure")
	}
	return "ok", nil
}
