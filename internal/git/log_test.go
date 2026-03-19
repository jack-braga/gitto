package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGetLog_RealRepo(t *testing.T) {
	// Create a temp repo with some commits
	tmp := t.TempDir()
	repoPath := filepath.Join(tmp, "test-repo")
	os.MkdirAll(repoPath, 0o755)

	// Init and make commits
	cmds := [][]string{
		{"git", "-C", repoPath, "init"},
		{"git", "-C", repoPath, "config", "user.email", "test@test.com"},
		{"git", "-C", repoPath, "config", "user.name", "Test"},
	}
	for _, c := range cmds {
		if err := exec.Command(c[0], c[1:]...).Run(); err != nil {
			t.Fatalf("setup command %v failed: %v", c, err)
		}
	}

	// Create files and commits
	os.WriteFile(filepath.Join(repoPath, "a.txt"), []byte("hello"), 0o644)
	exec.Command("git", "-C", repoPath, "add", ".").Run()
	exec.Command("git", "-C", repoPath, "commit", "-m", "first commit").Run()

	os.WriteFile(filepath.Join(repoPath, "b.txt"), []byte("world"), 0o644)
	exec.Command("git", "-C", repoPath, "add", ".").Run()
	exec.Command("git", "-C", repoPath, "commit", "-m", "second commit").Run()

	// Test GetLog
	entries, err := GetLog(repoPath, 10)
	if err != nil {
		t.Fatalf("GetLog failed: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	// Most recent first
	if entries[0].Message != "second commit" {
		t.Errorf("expected 'second commit', got %q", entries[0].Message)
	}
	if entries[1].Message != "first commit" {
		t.Errorf("expected 'first commit', got %q", entries[1].Message)
	}

	// Check fields are populated
	if entries[0].Hash == "" {
		t.Error("hash should not be empty")
	}
	if entries[0].FullHash == "" {
		t.Error("full hash should not be empty")
	}
	if entries[0].Author != "Test" {
		t.Errorf("expected author 'Test', got %q", entries[0].Author)
	}
	if entries[0].Date.IsZero() {
		t.Error("date should not be zero")
	}
}

func TestGetLog_EmptyRepo(t *testing.T) {
	tmp := t.TempDir()
	repoPath := filepath.Join(tmp, "empty-repo")
	os.MkdirAll(repoPath, 0o755)
	exec.Command("git", "-C", repoPath, "init").Run()

	// GetLog on empty repo should return error (git log fails with no commits)
	entries, err := GetLog(repoPath, 10)
	if err == nil && len(entries) == 0 {
		// This is acceptable — either an error or empty list
		return
	}
	if err != nil {
		// Also acceptable — git log fails on empty repo
		return
	}
	t.Errorf("unexpected: got %d entries with no error on empty repo", len(entries))
}
