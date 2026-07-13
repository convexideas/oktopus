package execution

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"

	"github.com/google/uuid"
)

// ProcessAdapter is a generic harness that wraps a CLI binary.
type ProcessAdapter struct {
	name       string
	binaryName string
	ArgsBuilder func(HarnessConfig) []string
}

// NewProcessAdapter creates a basic adapter. ArgsBuilder can be set for custom flag translation.
func NewProcessAdapter(name, binaryName string) *ProcessAdapter {
	return &ProcessAdapter{name: name, binaryName: binaryName}
}

func (a *ProcessAdapter) Name() string      { return a.name }
func (a *ProcessAdapter) BinaryName() string { return a.binaryName }

func (a *ProcessAdapter) Start(ctx context.Context, cfg HarnessConfig) (HarnessSession, error) {
	bin, err := exec.LookPath(a.binaryName)
	if err != nil {
		return nil, fmt.Errorf("%s CLI not found in PATH: %w", a.binaryName, err)
	}

	var args []string
	if a.ArgsBuilder != nil {
		args = a.ArgsBuilder(cfg)
	} else {
		args = cfg.Args
	}

	id := uuid.New().String()
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = cfg.Workspace
	cmd.Stdin = os.Stdin

	env := os.Environ()
	for k, v := range cfg.Env {
		env = append(env, k+"="+v)
	}
	if cfg.ProxyAddr != "" {
		env = append(env, "HTTPS_PROXY=http://"+cfg.ProxyAddr)
		env = append(env, "HTTP_PROXY=http://"+cfg.ProxyAddr)
	}
	cmd.Env = env

	var logFile *os.File
	var outputBuf *bytes.Buffer
	// Always capture stdout for memory
	outputBuf = &bytes.Buffer{}
	if cfg.LogPath != "" {
		logFile, err = os.Create(cfg.LogPath)
		if err != nil {
			return nil, fmt.Errorf("create log file: %w", err)
		}
		cmd.Stdout = io.MultiWriter(os.Stdout, logFile, outputBuf)
		cmd.Stderr = io.MultiWriter(os.Stderr, logFile)
	} else {
		cmd.Stdout = io.MultiWriter(os.Stdout, outputBuf)
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Start(); err != nil {
		if logFile != nil {
			logFile.Close()
		}
		return nil, fmt.Errorf("start %s: %w", a.name, err)
	}

	return &processSession{id: id, cmd: cmd, logFile: logFile, outputBuf: outputBuf}, nil
}

type processSession struct {
	id        string
	cmd       *exec.Cmd
	logFile   *os.File
	outputBuf *bytes.Buffer
	exitCode  int
}

func (s *processSession) ID() string { return s.id }

func (s *processSession) Wait() error {
	err := s.cmd.Wait()
	if s.logFile != nil {
		s.logFile.Close()
	}
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			s.exitCode = exitErr.ExitCode()
		} else {
			s.exitCode = 1
		}
		return err
	}
	return nil
}

func (s *processSession) Stop() error {
	if s.cmd.Process == nil {
		return nil
	}
	return s.cmd.Process.Signal(syscall.SIGTERM)
}

func (s *processSession) ExitCode() int { return s.exitCode }

// Output returns captured stdout content (non-interactive mode only).
func (s *processSession) Output() string {
	if s.outputBuf == nil {
		return ""
	}
	return s.outputBuf.String()
}
