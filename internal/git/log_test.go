package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestRepo creates a temp git repo with n commits and returns its path.
func setupTestRepo(t *testing.T, numCommits int) string {
	t.Helper()
	tmp := t.TempDir()
	repoPath := filepath.Join(tmp, "test-repo")
	os.MkdirAll(repoPath, 0o755)

	cmds := [][]string{
		{"git", "-C", repoPath, "init"},
		{"git", "-C", repoPath, "config", "user.email", "test@test.com"},
		{"git", "-C", repoPath, "config", "user.name", "Test User"},
	}
	for _, c := range cmds {
		if err := exec.Command(c[0], c[1:]...).Run(); err != nil {
			t.Fatalf("setup command %v failed: %v", c, err)
		}
	}

	for i := 0; i < numCommits; i++ {
		fname := filepath.Join(repoPath, string(rune('a'+i))+".txt")
		os.WriteFile(fname, []byte("content"), 0o644)
		exec.Command("git", "-C", repoPath, "add", ".").Run()
		exec.Command("git", "-C", repoPath, "commit", "-m", "commit "+string(rune('A'+i))).Run()
	}

	return repoPath
}

func TestGetLog_RealRepo(t *testing.T) {
	repoPath := setupTestRepo(t, 2)

	entries, err := GetLog(repoPath, 10)
	if err != nil {
		t.Fatalf("GetLog failed: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	// Most recent first
	if entries[0].Message != "commit B" {
		t.Errorf("expected 'commit B', got %q", entries[0].Message)
	}
	if entries[1].Message != "commit A" {
		t.Errorf("expected 'commit A', got %q", entries[1].Message)
	}

	// Check fields are populated
	if entries[0].Hash == "" {
		t.Error("hash should not be empty")
	}
	if entries[0].FullHash == "" {
		t.Error("full hash should not be empty")
	}
	if entries[0].Author != "Test User" {
		t.Errorf("expected author 'Test User', got %q", entries[0].Author)
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

	// GetLog on empty repo — either error or empty list is acceptable
	entries, err := GetLog(repoPath, 10)
	if err == nil && len(entries) > 0 {
		t.Errorf("unexpected: got %d entries on empty repo", len(entries))
	}
}

func TestGetLog_LimitRespected(t *testing.T) {
	repoPath := setupTestRepo(t, 5)

	entries, err := GetLog(repoPath, 3)
	if err != nil {
		t.Fatalf("GetLog failed: %v", err)
	}

	if len(entries) != 3 {
		t.Fatalf("expected 3 entries (limit), got %d", len(entries))
	}
}

func TestGetLog_FieldsNotCorrupted(t *testing.T) {
	// Regression test: ensure the log separator doesn't appear in parsed fields.
	// The original bug used NUL bytes which broke exec on macOS.
	repoPath := setupTestRepo(t, 1)

	entries, err := GetLog(repoPath, 10)
	if err != nil {
		t.Fatalf("GetLog failed: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least 1 entry")
	}

	e := entries[0]

	// No field should contain the separator character
	for _, field := range []string{e.Hash, e.FullHash, e.Message, e.Author} {
		if strings.Contains(field, logSeparator) {
			t.Errorf("field contains separator: %q", field)
		}
	}

	// Hash should be 7 chars, FullHash 40
	if len(e.Hash) < 7 {
		t.Errorf("short hash too short: %q", e.Hash)
	}
	if len(e.FullHash) != 40 {
		t.Errorf("full hash should be 40 chars, got %d: %q", len(e.FullHash), e.FullHash)
	}
}

func TestGetLog_SpecialCharsInMessage(t *testing.T) {
	// Ensure commit messages with special chars don't break parsing
	tmp := t.TempDir()
	repoPath := filepath.Join(tmp, "special-repo")
	os.MkdirAll(repoPath, 0o755)

	exec.Command("git", "-C", repoPath, "init").Run()
	exec.Command("git", "-C", repoPath, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", repoPath, "config", "user.name", "Test").Run()

	os.WriteFile(filepath.Join(repoPath, "a.txt"), []byte("x"), 0o644)
	exec.Command("git", "-C", repoPath, "add", ".").Run()
	// Message with tabs, quotes, colons, parens — typical conventional commit
	exec.Command("git", "-C", repoPath, "commit", "-m", `feat(api): add "search" endpoint — fixes #42`).Run()

	entries, err := GetLog(repoPath, 10)
	if err != nil {
		t.Fatalf("GetLog failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Message != `feat(api): add "search" endpoint — fixes #42` {
		t.Errorf("message mangled: %q", entries[0].Message)
	}
}

func TestGetLog_Refs(t *testing.T) {
	repoPath := setupTestRepo(t, 1)

	entries, err := GetLog(repoPath, 10)
	if err != nil {
		t.Fatalf("GetLog failed: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least 1 entry")
	}

	// HEAD commit should have refs (at minimum HEAD -> main/master)
	if len(entries[0].Refs) == 0 {
		t.Error("HEAD commit should have at least one ref")
	}
}

func TestLogSeparator_NotNUL(t *testing.T) {
	// Regression guard: NUL bytes in exec args cause "invalid argument" on macOS.
	// The separator must never be \x00.
	if logSeparator == "\x00" {
		t.Fatal("logSeparator must not be NUL — breaks exec.Command on macOS")
	}
}
