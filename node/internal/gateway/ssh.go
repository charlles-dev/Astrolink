package gateway

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHConfig struct {
	Host    string
	Port    int
	User    string
	KeyPath string
	Timeout time.Duration
}

type SSHRunner struct {
	cfg SSHConfig
}

func NewSSHRunner(cfg SSHConfig) *SSHRunner {
	if cfg.Port == 0 {
		cfg.Port = 22
	}
	if cfg.User == "" {
		cfg.User = "root"
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 10 * time.Second
	}
	return &SSHRunner{cfg: cfg}
}

func (r *SSHRunner) Run(ctx context.Context, command string) (string, error) {
	if r.cfg.Host == "" {
		return "", errors.New("host SSH do OpenNDS nao configurado")
	}
	if r.cfg.KeyPath == "" {
		return "", errors.New("chave SSH do OpenNDS nao configurada")
	}

	key, err := os.ReadFile(r.cfg.KeyPath)
	if err != nil {
		return "", fmt.Errorf("ler chave SSH: %w", err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return "", fmt.Errorf("parse chave SSH: %w", err)
	}

	address := net.JoinHostPort(r.cfg.Host, strconv.Itoa(r.cfg.Port))
	dialer := net.Dialer{Timeout: r.cfg.Timeout}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return "", fmt.Errorf("conectar SSH %s: %w", address, err)
	}

	clientConfig := &ssh.ClientConfig{
		User:            r.cfg.User,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         r.cfg.Timeout,
	}
	sshConn, chans, reqs, err := ssh.NewClientConn(conn, address, clientConfig)
	if err != nil {
		_ = conn.Close()
		return "", fmt.Errorf("iniciar SSH %s: %w", address, err)
	}
	client := ssh.NewClient(sshConn, chans, reqs)
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("criar sessao SSH: %w", err)
	}
	defer session.Close()

	type runResult struct {
		output []byte
		err    error
	}
	done := make(chan runResult, 1)
	go func() {
		output, err := session.CombinedOutput(command)
		done <- runResult{output: output, err: err}
	}()

	select {
	case <-ctx.Done():
		_ = session.Close()
		return "", ctx.Err()
	case result := <-done:
		if result.err != nil {
			return string(result.output), fmt.Errorf("executar comando SSH %q: %w: %s", command, result.err, string(result.output))
		}
		return string(result.output), nil
	}
}
