package process

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

// OutputHandler is called for each line of output from the process
type OutputHandler func(line string, isStderr bool)

// Manager manages the child process
type Manager struct {
	cmd           *exec.Cmd
	proxyPort     int
	outputHandler OutputHandler
	mu            sync.Mutex
	started       bool
	ctx           context.Context
	cancel        context.CancelFunc
}

// Config holds process manager configuration
type Config struct {
	Command       []string
	ProxyPort     int
	OutputHandler OutputHandler
}

// New creates a new process Manager
func New(cfg Config) (*Manager, error) {
	if len(cfg.Command) == 0 {
		return nil, fmt.Errorf("no command specified")
	}

	ctx, cancel := context.WithCancel(context.Background())

	m := &Manager{
		proxyPort:     cfg.ProxyPort,
		outputHandler: cfg.OutputHandler,
		ctx:           ctx,
		cancel:        cancel,
	}

	// Create the command
	if len(cfg.Command) == 1 {
		m.cmd = exec.CommandContext(ctx, cfg.Command[0])
	} else {
		m.cmd = exec.CommandContext(ctx, cfg.Command[0], cfg.Command[1:]...)
	}

	return m, nil
}

// Start starts the child process with proxy environment variables
func (m *Manager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.started {
		return fmt.Errorf("process already started")
	}

	// Set up environment with proxy
	m.cmd.Env = m.buildEnv()

	// Set up pipes for stdout and stderr
	stdout, err := m.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := m.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Connect stdin
	m.cmd.Stdin = os.Stdin

	// Start the process
	if err := m.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}

	m.started = true

	// Handle output in goroutines
	go m.handleOutput(stdout, false)
	go m.handleOutput(stderr, true)

	return nil
}

// buildEnv creates the environment variables for the child process
func (m *Manager) buildEnv() []string {
	env := os.Environ()
	proxyURL := fmt.Sprintf("http://127.0.0.1:%d", m.proxyPort)

	// Add/override proxy environment variables
	proxyVars := map[string]string{
		"HTTP_PROXY":  proxyURL,
		"http_proxy":  proxyURL,
		"HTTPS_PROXY": proxyURL,
		"https_proxy": proxyURL,
		// Force proxy for localhost (many clients skip localhost by default)
		"NO_PROXY":  "",
		"no_proxy":  "",
		// A2A specific - some implementations use these
		"A2A_PROXY":    proxyURL,
		"A2A_TRACE":    "1",
		"A2A_TRACE_UI": fmt.Sprintf("http://127.0.0.1:%d/ui", m.proxyPort),
	}

	// Remove existing proxy vars and add new ones
	filteredEnv := make([]string, 0, len(env)+len(proxyVars))
	for _, e := range env {
		key := strings.Split(e, "=")[0]
		if _, isProxy := proxyVars[key]; !isProxy {
			filteredEnv = append(filteredEnv, e)
		}
	}

	for key, value := range proxyVars {
		filteredEnv = append(filteredEnv, fmt.Sprintf("%s=%s", key, value))
	}

	return filteredEnv
}

// handleOutput reads from a pipe and calls the output handler
func (m *Manager) handleOutput(pipe io.ReadCloser, isStderr bool) {
	scanner := bufio.NewScanner(pipe)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB buffer

	for scanner.Scan() {
		line := scanner.Text()
		
		// Always print to appropriate output
		if isStderr {
			fmt.Fprintln(os.Stderr, line)
		} else {
			fmt.Println(line)
		}

		// Call handler if set
		if m.outputHandler != nil {
			m.outputHandler(line, isStderr)
		}
	}
}

// Wait waits for the process to exit and returns the exit code
func (m *Manager) Wait() (int, error) {
	if m.cmd == nil || m.cmd.Process == nil {
		return -1, fmt.Errorf("process not started")
	}

	err := m.cmd.Wait()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode(), nil
		}
		return -1, err
	}

	return 0, nil
}

// Stop stops the child process gracefully
func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cmd == nil || m.cmd.Process == nil {
		return nil
	}

	m.cancel()

	// Try graceful shutdown first (SIGTERM)
	if err := m.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		// Process might have already exited
		return nil
	}

	return nil
}

// Kill forcefully kills the child process
func (m *Manager) Kill() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cmd == nil || m.cmd.Process == nil {
		return nil
	}

	return m.cmd.Process.Kill()
}

// PID returns the process ID of the child process
func (m *Manager) PID() int {
	if m.cmd == nil || m.cmd.Process == nil {
		return -1
	}
	return m.cmd.Process.Pid
}

// IsRunning returns true if the process is still running
func (m *Manager) IsRunning() bool {
	if m.cmd == nil || m.cmd.Process == nil {
		return false
	}
	
	// Check if process is still running
	err := m.cmd.Process.Signal(syscall.Signal(0))
	return err == nil
}

// SetupSignalHandling sets up signal handling for graceful shutdown
func (m *Manager) SetupSignalHandling() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		fmt.Printf("\n📍 Received signal %v, shutting down...\n", sig)
		_ = m.Stop()
	}()
}

// CommandString returns the command as a string
func (m *Manager) CommandString() string {
	if m.cmd == nil {
		return ""
	}
	return strings.Join(m.cmd.Args, " ")
}

