package git

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// GitError represents an error from a git command.
type GitError struct {
	Command []string
	Stderr  string
	Err     error
}

// Error returns a human-readable error string.
func (e *GitError) Error() string {
	stderr := strings.TrimSpace(e.Stderr)
	if stderr != "" {
		return fmt.Sprintf("git %s: %s", strings.Join(e.Command, " "), stderr)
	}
	return fmt.Sprintf("git %s: %v", strings.Join(e.Command, " "), e.Err)
}

// Unwrap returns the underlying error.
func (e *GitError) Unwrap() error {
	return e.Err
}

// GitExec runs a git command in the given directory with a 30-second timeout.
func GitExec(repoPath string, args ...string) (string, error) {
	return gitExecWithTimeout(repoPath, 30*time.Second, args...)
}

// GitExecLong runs a git command with a 120-second timeout, for push/pull/fetch.
func GitExecLong(repoPath string, args ...string) (string, error) {
	return gitExecWithTimeout(repoPath, 120*time.Second, args...)
}

func gitExecWithTimeout(repoPath string, timeout time.Duration, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", repoPath}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", &GitError{
			Command: args,
			Stderr:  stderr.String(),
			Err:     err,
		}
	}

	return stdout.String(), nil
}
