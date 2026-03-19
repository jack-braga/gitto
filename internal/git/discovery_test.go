package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverRepos(t *testing.T) {
	tmp := t.TempDir()

	// Create some child directories
	repos := []string{"alpha", "beta", "gamma"}
	for _, name := range repos {
		dir := filepath.Join(tmp, name)
		os.MkdirAll(filepath.Join(dir, ".git"), 0o755)
	}

	// Create a non-repo directory
	os.MkdirAll(filepath.Join(tmp, "not-a-repo"), 0o755)

	// Create a file (should be ignored)
	os.WriteFile(filepath.Join(tmp, "file.txt"), []byte("hello"), 0o644)

	found, err := DiscoverRepos(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(found) != 3 {
		t.Fatalf("expected 3 repos, got %d: %v", len(found), found)
	}

	// Should be sorted
	for i, name := range repos {
		expected := filepath.Join(tmp, name)
		if found[i] != expected {
			t.Errorf("expected %s at index %d, got %s", expected, i, found[i])
		}
	}
}

func TestDiscoverRepos_Empty(t *testing.T) {
	tmp := t.TempDir()

	found, err := DiscoverRepos(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(found) != 0 {
		t.Errorf("expected 0 repos, got %d", len(found))
	}
}

func TestDiscoverRepos_GitFile(t *testing.T) {
	// Test that .git as a file (worktree) is also detected
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "worktree")
	os.MkdirAll(dir, 0o755)
	// .git as a file (like in worktrees/submodules)
	os.WriteFile(filepath.Join(dir, ".git"), []byte("gitdir: /somewhere/else"), 0o644)

	found, err := DiscoverRepos(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(found) != 1 {
		t.Fatalf("expected 1 repo (worktree), got %d", len(found))
	}
}
