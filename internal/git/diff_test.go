package git

import "testing"

func TestParseNumstat(t *testing.T) {
	input := `10	3	src/main.go
5	0	src/new.go
0	12	src/removed.go
`
	stats := parseNumstat(input)

	if len(stats) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(stats))
	}

	main := stats["src/main.go"]
	if main.Insertions != 10 || main.Deletions != 3 {
		t.Errorf("main.go: got +%d -%d, want +10 -3", main.Insertions, main.Deletions)
	}

	newFile := stats["src/new.go"]
	if newFile.Insertions != 5 || newFile.Deletions != 0 {
		t.Errorf("new.go: got +%d -%d, want +5 -0", newFile.Insertions, newFile.Deletions)
	}

	removed := stats["src/removed.go"]
	if removed.Insertions != 0 || removed.Deletions != 12 {
		t.Errorf("removed.go: got +%d -%d, want +0 -12", removed.Insertions, removed.Deletions)
	}
}

func TestParseNumstat_Empty(t *testing.T) {
	stats := parseNumstat("")
	if len(stats) != 0 {
		t.Errorf("expected 0 entries for empty input, got %d", len(stats))
	}
}

func TestParseNumstat_Binary(t *testing.T) {
	// Binary files show - for both
	input := "-\t-\timage.png\n"
	stats := parseNumstat(input)
	if len(stats) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(stats))
	}
	png := stats["image.png"]
	if png.Insertions != 0 || png.Deletions != 0 {
		t.Errorf("binary: got +%d -%d, want +0 -0", png.Insertions, png.Deletions)
	}
}
