package styles

import (
	"testing"
	"time"

	"github.com/jack-braga/gitto/internal/git"
)

func TestStatusChar(t *testing.T) {
	tests := []struct {
		status git.FileStatus
		want   string
	}{
		{git.StatusModified, "M"},
		{git.StatusAdded, "A"},
		{git.StatusDeleted, "D"},
		{git.StatusRenamed, "R"},
		{git.StatusCopied, "C"},
		{git.StatusUntracked, "?"},
		{git.StatusConflicted, "U"},
	}

	for _, tt := range tests {
		got := StatusChar(tt.status)
		if got != tt.want {
			t.Errorf("StatusChar(%d) = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxWidth int
		want     string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "hell…"},
		{"hi", 2, "hi"},
		{"hi", 1, "…"},
		{"", 5, ""},
		{"abc", 0, ""},
	}

	for _, tt := range tests {
		got := Truncate(tt.input, tt.maxWidth)
		if got != tt.want {
			t.Errorf("Truncate(%q, %d) = %q, want %q", tt.input, tt.maxWidth, got, tt.want)
		}
	}
}

func TestRelativeTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		input time.Time
		want  string
	}{
		{time.Time{}, ""},
		{now.Add(-30 * time.Second), "just now"},
		{now.Add(-5 * time.Minute), "5m ago"},
		{now.Add(-3 * time.Hour), "3h ago"},
		{now.Add(-2 * 24 * time.Hour), "2d ago"},
		{now.Add(-2 * 7 * 24 * time.Hour), "2w ago"},
	}

	for _, tt := range tests {
		got := RelativeTime(tt.input)
		if got != tt.want {
			t.Errorf("RelativeTime(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestContentHeight(t *testing.T) {
	// ChromeHeight is 6
	if h := ContentHeight(30); h != 24 {
		t.Errorf("ContentHeight(30) = %d, want 24", h)
	}
	// Very small terminal
	if h := ContentHeight(5); h != 1 {
		t.Errorf("ContentHeight(5) = %d, want 1", h)
	}
}

func TestDivider(t *testing.T) {
	d := Divider(10)
	if d == "" {
		t.Error("Divider(10) returned empty string")
	}
}

func TestPadOrTruncate(t *testing.T) {
	// Pad
	got := PadOrTruncate("hi", 5)
	if len(got) != 5 {
		t.Errorf("PadOrTruncate(\"hi\", 5) length = %d, want 5", len(got))
	}
	// Truncate
	got = PadOrTruncate("hello world", 5)
	if got != "hell…" {
		t.Errorf("PadOrTruncate(\"hello world\", 5) = %q, want \"hell…\"", got)
	}
}
