package git

import (
	"testing"
)

func TestParsePorcelainV2_BranchInfo(t *testing.T) {
	input := `# branch.oid abc123def456
# branch.head main
# branch.upstream origin/main
# branch.ab +3 -1
`
	repo, err := ParsePorcelainV2(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.Branch != "main" {
		t.Errorf("expected branch=main, got %q", repo.Branch)
	}
	if repo.Upstream != "origin/main" {
		t.Errorf("expected upstream=origin/main, got %q", repo.Upstream)
	}
	if repo.Ahead != 3 {
		t.Errorf("expected ahead=3, got %d", repo.Ahead)
	}
	if repo.Behind != 1 {
		t.Errorf("expected behind=1, got %d", repo.Behind)
	}
}

func TestParsePorcelainV2_OrdinaryChanges(t *testing.T) {
	input := `# branch.oid abc123
# branch.head feat
1 M. N... 100644 100644 100644 abc123 def456 src/main.go
1 .M N... 100644 100644 100644 abc123 def456 src/utils.go
1 MM N... 100644 100644 100644 abc123 def456 src/both.go
`
	repo, err := ParsePorcelainV2(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// src/main.go: staged modified (M.), src/both.go: staged (M)
	if len(repo.Staged) != 2 {
		t.Errorf("expected 2 staged, got %d", len(repo.Staged))
	}

	// src/utils.go: unstaged modified (.M), src/both.go: unstaged (M)
	if len(repo.Unstaged) != 2 {
		t.Errorf("expected 2 unstaged, got %d", len(repo.Unstaged))
	}
}

func TestParsePorcelainV2_Untracked(t *testing.T) {
	input := `# branch.oid abc123
# branch.head main
? newfile.txt
? another/new.go
`
	repo, err := ParsePorcelainV2(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(repo.Untracked) != 2 {
		t.Errorf("expected 2 untracked, got %d", len(repo.Untracked))
	}
	if repo.Untracked[0].Path != "newfile.txt" {
		t.Errorf("expected newfile.txt, got %q", repo.Untracked[0].Path)
	}
	if repo.Untracked[0].Status != StatusUntracked {
		t.Errorf("expected StatusUntracked, got %d", repo.Untracked[0].Status)
	}
}

func TestParsePorcelainV2_Rename(t *testing.T) {
	input := `# branch.oid abc123
# branch.head main
2 R. N... 100644 100644 100644 abc123 def456 R100 new.go	old.go
`
	repo, err := ParsePorcelainV2(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(repo.Staged) != 1 {
		t.Fatalf("expected 1 staged rename, got %d", len(repo.Staged))
	}
	if repo.Staged[0].Status != StatusRenamed {
		t.Errorf("expected StatusRenamed, got %d", repo.Staged[0].Status)
	}
	if repo.Staged[0].Path != "new.go" {
		t.Errorf("expected new.go, got %q", repo.Staged[0].Path)
	}
	if repo.Staged[0].OldPath != "old.go" {
		t.Errorf("expected old.go, got %q", repo.Staged[0].OldPath)
	}
}

func TestParsePorcelainV2_Unmerged(t *testing.T) {
	input := `# branch.oid abc123
# branch.head main
u UU N... 100644 100644 100644 100644 abc123 def456 ghi789 conflict.go
`
	repo, err := ParsePorcelainV2(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(repo.Unstaged) != 1 {
		t.Fatalf("expected 1 conflicted, got %d", len(repo.Unstaged))
	}
	if repo.Unstaged[0].Status != StatusConflicted {
		t.Errorf("expected StatusConflicted, got %d", repo.Unstaged[0].Status)
	}
}

func TestParsePorcelainV2_Empty(t *testing.T) {
	input := `# branch.oid abc123
# branch.head main
`
	repo, err := ParsePorcelainV2(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !repo.IsClean() {
		t.Error("expected clean repo")
	}
	if repo.TotalChanges() != 0 {
		t.Errorf("expected 0 changes, got %d", repo.TotalChanges())
	}
}

func TestParsePorcelainV2_AddedFile(t *testing.T) {
	input := `# branch.oid abc123
# branch.head main
1 A. N... 000000 100644 100644 0000000 abc1234 newfile.go
`
	repo, err := ParsePorcelainV2(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(repo.Staged) != 1 {
		t.Fatalf("expected 1 staged, got %d", len(repo.Staged))
	}
	if repo.Staged[0].Status != StatusAdded {
		t.Errorf("expected StatusAdded, got %d", repo.Staged[0].Status)
	}
}

func TestParsePorcelainV2_DeletedFile(t *testing.T) {
	input := `# branch.oid abc123
# branch.head main
1 D. N... 100644 000000 000000 abc1234 0000000 removed.go
`
	repo, err := ParsePorcelainV2(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(repo.Staged) != 1 {
		t.Fatalf("expected 1 staged, got %d", len(repo.Staged))
	}
	if repo.Staged[0].Status != StatusDeleted {
		t.Errorf("expected StatusDeleted, got %d", repo.Staged[0].Status)
	}
}
