package cli

import (
	"os"
	"os/exec"
	"strings"
)

// defaultPager is used when the PAGER env var is unset. `-R` preserves ANSI
// color sequences and `-F` quits if output fits on one screen.
const defaultPager = "less -RF"

// pagerProcess tracks an active pager subprocess and the original stdout so it
// can be restored when the pager exits.
type pagerProcess struct {
	cmd            *exec.Cmd
	pipeWriter     *os.File
	originalStdout *os.File
	stopped        bool
}

// startPager spawns a pager subprocess and redirects os.Stdout to it. Returns
// nil if no pager is configured, stdout is not a TTY, or spawning fails.
// Callers must invoke stop() before the program exits.
func startPager() *pagerProcess {
	if !IsTTY() {
		return nil
	}

	pagerCmd := strings.TrimSpace(os.Getenv("PAGER"))
	if pagerCmd == "" {
		pagerCmd = defaultPager
	}

	parts := strings.Fields(pagerCmd)
	if len(parts) == 0 {
		return nil
	}

	r, w, err := os.Pipe()
	if err != nil {
		return nil
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdin = r
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		_ = r.Close()
		_ = w.Close()
		return nil
	}

	// The pager owns the read end now.
	_ = r.Close()

	original := os.Stdout
	os.Stdout = w

	return &pagerProcess{
		cmd:            cmd,
		pipeWriter:     w,
		originalStdout: original,
	}
}

// stop closes the write end of the pipe (causing the pager to exit when it
// finishes displaying) and waits for the pager to finish before returning.
// Safe to call multiple times.
func (p *pagerProcess) stop() {
	if p == nil || p.stopped {
		return
	}
	p.stopped = true
	_ = p.pipeWriter.Close()
	os.Stdout = p.originalStdout
	_ = p.cmd.Wait()
}
