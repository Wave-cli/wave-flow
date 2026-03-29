// Package flow implements the core logic for the wave-flow plugin.
package flow

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

// Watcher watches files and restarts a command on changes.
type Watcher struct {
	cmd      *Command
	stdout   io.Writer
	stderr   io.Writer
	process  *exec.Cmd
	mu       sync.Mutex
	stopCh   chan struct{}
	doneCh   chan struct{}
	debounce time.Duration
}

// NewWatcher creates a new watcher for the given command.
func NewWatcher(cmd *Command, stdout, stderr io.Writer) *Watcher {
	return &Watcher{
		cmd:      cmd,
		stdout:   stdout,
		stderr:   stderr,
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
		debounce: 500 * time.Millisecond,
	}
}

// Start begins watching files and running the command.
// It blocks until Stop is called or an error occurs.
func (w *Watcher) Start() error {
	if len(w.cmd.Watch) == 0 {
		return fmt.Errorf("no watch patterns specified")
	}

	// Start the command initially
	if err := w.startProcess(); err != nil {
		return fmt.Errorf("starting command: %w", err)
	}

	fmt.Fprintf(w.stdout, "Watching: %v\n", w.cmd.Watch)

	// Start the watch loop
	go w.watchLoop()

	// Wait for stop signal
	<-w.stopCh
	close(w.doneCh)
	return nil
}

// Stop stops the watcher and kills any running process.
func (w *Watcher) Stop() {
	close(w.stopCh)
	w.killProcess()
	<-w.doneCh
}

// startProcess starts the command process.
func (w *Watcher) startProcess() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	c := newShellCmd(w.cmd.Cmd, w.cmd.Env, w.stdout, w.stderr)

	// Create new process group so we can kill children too
	c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := c.Start(); err != nil {
		return err
	}

	w.process = c

	// Wait for process in background
	go func() {
		c.Wait()
	}()

	return nil
}

// killProcess kills the running process and its children.
func (w *Watcher) killProcess() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.process == nil || w.process.Process == nil {
		return
	}

	// Kill the process group (negative PID)
	pgid, err := syscall.Getpgid(w.process.Process.Pid)
	if err == nil {
		syscall.Kill(-pgid, syscall.SIGTERM)
	}

	// Give it a moment to exit gracefully
	done := make(chan struct{})
	go func() {
		w.process.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Process exited gracefully
	case <-time.After(2 * time.Second):
		// Force kill if it doesn't exit
		if pgid != 0 {
			syscall.Kill(-pgid, syscall.SIGKILL)
		}
	}

	w.process = nil
}

// restart kills the current process and starts a new one.
func (w *Watcher) restart() {
	fmt.Fprintln(w.stdout, "\n--- Restarting ---")
	w.killProcess()
	if err := w.startProcess(); err != nil {
		fmt.Fprintf(w.stderr, "Error restarting: %v\n", err)
	}
}

// watchLoop polls files for changes and triggers restarts.
func (w *Watcher) watchLoop() {
	// Build initial file state
	files := make(map[string]time.Time)
	w.scanFiles(files)

	ticker := time.NewTicker(w.debounce)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			if w.checkForChanges(files) {
				w.restart()
			}
		}
	}
}

// scanFiles finds all files matching watch patterns and records their mod times.
func (w *Watcher) scanFiles(files map[string]time.Time) {
	for _, pattern := range w.cmd.Watch {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil {
				continue
			}
			if !info.IsDir() {
				files[match] = info.ModTime()
			}
		}
	}
}

// checkForChanges checks if any watched files have changed.
// Returns true if changes detected, and updates the files map.
func (w *Watcher) checkForChanges(files map[string]time.Time) bool {
	changed := false

	// Check existing files and look for new ones
	currentFiles := make(map[string]time.Time)
	for _, pattern := range w.cmd.Watch {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil {
				continue
			}
			if info.IsDir() {
				continue
			}

			currentFiles[match] = info.ModTime()

			// Check if file is new or modified
			if oldTime, exists := files[match]; !exists || !oldTime.Equal(info.ModTime()) {
				changed = true
			}
		}
	}

	// Check for deleted files
	for path := range files {
		if _, exists := currentFiles[path]; !exists {
			changed = true
		}
	}

	// Update files map
	for k := range files {
		delete(files, k)
	}
	for k, v := range currentFiles {
		files[k] = v
	}

	return changed
}

// RunWithWatch runs a command with watch mode enabled.
// It watches the specified file patterns and restarts on changes.
func RunWithWatch(cmd *Command, stdout, stderr io.Writer) int {
	watcher := NewWatcher(cmd, stdout, stderr)

	// Handle interrupt signal
	sigCh := make(chan os.Signal, 1)
	go func() {
		<-sigCh
		fmt.Fprintln(stdout, "\nStopping watch mode...")
		watcher.Stop()
	}()

	// Start watching
	if err := watcher.Start(); err != nil {
		fmt.Fprintf(stderr, "Watch error: %v\n", err)
		return 1
	}

	return 0
}
