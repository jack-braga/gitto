package git

import "testing"

func TestRepository_IsClean(t *testing.T) {
	repo := &Repository{}
	if !repo.IsClean() {
		t.Error("empty repo should be clean")
	}
	if repo.TotalChanges() != 0 {
		t.Errorf("expected 0 changes, got %d", repo.TotalChanges())
	}
}

func TestRepository_TotalChanges(t *testing.T) {
	repo := &Repository{
		Staged:    []FileChange{{Path: "a.go"}},
		Unstaged:  []FileChange{{Path: "b.go"}, {Path: "c.go"}},
		Untracked: []FileChange{{Path: "d.go"}},
	}
	if repo.TotalChanges() != 4 {
		t.Errorf("expected 4 changes, got %d", repo.TotalChanges())
	}
	if repo.IsClean() {
		t.Error("repo with changes should not be clean")
	}
}

func TestRepository_TotalChanges_StagedOnly(t *testing.T) {
	repo := &Repository{
		Staged: []FileChange{{Path: "a.go"}, {Path: "b.go"}},
	}
	if repo.TotalChanges() != 2 {
		t.Errorf("expected 2 changes, got %d", repo.TotalChanges())
	}
}
