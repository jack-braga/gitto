package styles

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/jack-braga/gitto/internal/git"
)

// Minimum terminal dimensions before showing a resize warning.
const (
	MinWidth  = 40
	MinHeight = 10
)

// Fixed layout heights (lines consumed by chrome).
const (
	HeaderHeight  = 1
	DividerHeight = 1
	TabsHeight    = 1
	FooterHeight  = 1
	// Total chrome = header + divider + tabs + divider + divider + footer = 6
	ChromeHeight = HeaderHeight + DividerHeight + TabsHeight + DividerHeight + DividerHeight + FooterHeight
)

// Colors — adaptive for light and dark terminals.
var (
	Subtle    = lipgloss.AdaptiveColor{Light: "#888888", Dark: "#6c7086"}
	Highlight = lipgloss.AdaptiveColor{Light: "#1a73e8", Dark: "#89b4fa"}
	Text      = lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#cdd6f4"}
	Modified  = lipgloss.AdaptiveColor{Light: "#b87800", Dark: "#f9e2af"}
	Added     = lipgloss.AdaptiveColor{Light: "#1e7d34", Dark: "#a6e3a1"}
	Deleted   = lipgloss.AdaptiveColor{Light: "#d32f2f", Dark: "#f38ba8"}
	Conflict  = lipgloss.AdaptiveColor{Light: "#d32f2f", Dark: "#f38ba8"}
	Border    = lipgloss.AdaptiveColor{Light: "#cccccc", Dark: "#45475a"}
	Surface   = lipgloss.AdaptiveColor{Light: "#f5f5f5", Dark: "#313244"}
	Clean     = Added
)

// Component styles.
var (
	HeaderStyle     = lipgloss.NewStyle().Bold(true).Foreground(Highlight).Padding(0, 1)
	HeaderPathStyle = lipgloss.NewStyle().Foreground(Subtle)

	BranchPill = lipgloss.NewStyle().
			Background(Surface).
			Foreground(Highlight).
			Padding(0, 1)

	TabActiveStyle   = lipgloss.NewStyle().Bold(true).Foreground(Highlight).Underline(true)
	TabInactiveStyle = lipgloss.NewStyle().Foreground(Subtle)
	ViewLabelStyle   = lipgloss.NewStyle().Foreground(Subtle).Italic(true)

	RepoNameStyle        = lipgloss.NewStyle().Bold(true).Foreground(Text)
	RepoNameFocusedStyle = lipgloss.NewStyle().Bold(true).Foreground(Highlight)
	CleanStyle           = lipgloss.NewStyle().Foreground(Clean)
	ModifiedStyle        = lipgloss.NewStyle().Foreground(Modified)
	AddedStyle           = lipgloss.NewStyle().Foreground(Added)
	DeletedStyle         = lipgloss.NewStyle().Foreground(Deleted)
	ConflictStyle        = lipgloss.NewStyle().Foreground(Conflict)
	UntrackedStyle       = lipgloss.NewStyle().Foreground(Subtle)

	StagedHeaderStyle   = lipgloss.NewStyle().Bold(true).Foreground(Added)
	UnstagedHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(Modified)
	StashStyle          = lipgloss.NewStyle().Foreground(Subtle)

	FooterStyle    = lipgloss.NewStyle().Foreground(Subtle)
	FooterKeyStyle = lipgloss.NewStyle().Bold(true).Foreground(Highlight)

	SelectedStyle = lipgloss.NewStyle().Background(Surface)
	DividerStyle  = lipgloss.NewStyle().Foreground(Border)

	CommitInputStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Border).
				Padding(0, 1)
	CommitLabelStyle = lipgloss.NewStyle().Foreground(Subtle).Bold(true)

	AheadStyle  = lipgloss.NewStyle().Foreground(Added)
	BehindStyle = lipgloss.NewStyle().Foreground(Deleted)

	HashStyle   = lipgloss.NewStyle().Foreground(Highlight)
	AuthorStyle = lipgloss.NewStyle().Foreground(Subtle)
	DateStyle   = lipgloss.NewStyle().Foreground(Subtle)
	RefStyle    = lipgloss.NewStyle().Background(Surface).Foreground(Highlight).Padding(0, 1)
)

// Icons.
const (
	ExpandedIcon  = "\u25bc" // ▼
	CollapsedIcon = "\u25ba" // ►
	FileIndent    = "  "
)

// StatusStyle returns the appropriate style for a git file status.
func StatusStyle(status git.FileStatus) lipgloss.Style {
	switch status {
	case git.StatusModified:
		return ModifiedStyle
	case git.StatusAdded:
		return AddedStyle
	case git.StatusDeleted:
		return DeletedStyle
	case git.StatusRenamed:
		return ModifiedStyle
	case git.StatusCopied:
		return AddedStyle
	case git.StatusUntracked:
		return UntrackedStyle
	case git.StatusConflicted:
		return ConflictStyle
	default:
		return lipgloss.NewStyle()
	}
}

// StatusChar returns the single-character representation of a file status.
func StatusChar(status git.FileStatus) string {
	switch status {
	case git.StatusModified:
		return "M"
	case git.StatusAdded:
		return "A"
	case git.StatusDeleted:
		return "D"
	case git.StatusRenamed:
		return "R"
	case git.StatusCopied:
		return "C"
	case git.StatusUntracked:
		return "?"
	case git.StatusConflicted:
		return "U"
	default:
		return " "
	}
}

// Divider renders a horizontal line.
func Divider(width int) string {
	if width <= 0 {
		width = 40
	}
	return DividerStyle.Render(strings.Repeat("─", width))
}

// Truncate truncates a string to maxWidth, appending "…" if truncated.
func Truncate(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if len(s) <= maxWidth {
		return s
	}
	if maxWidth <= 1 {
		return "…"
	}
	return s[:maxWidth-1] + "…"
}

// ClampLine ensures a rendered line doesn't exceed maxWidth by truncating.
// It operates on the raw string (may break ANSI — use on pre-render text).
func ClampLine(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	// lipgloss.Width counts visible runes, ignoring ANSI
	w := lipgloss.Width(s)
	if w <= maxWidth {
		return s
	}
	// Rough truncation — trim runes from end
	runes := []rune(s)
	for lipgloss.Width(string(runes)) > maxWidth && len(runes) > 0 {
		runes = runes[:len(runes)-1]
	}
	return string(runes)
}

// PadOrTruncate pads a string to exactly width, or truncates with "…".
func PadOrTruncate(s string, width int) string {
	w := lipgloss.Width(s)
	if w == width {
		return s
	}
	if w > width {
		return Truncate(s, width)
	}
	return s + strings.Repeat(" ", width-w)
}

// RelativeTime returns a human-readable relative time string.
func RelativeTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	d := time.Since(t)
	seconds := int(math.Abs(d.Seconds()))

	switch {
	case seconds < 60:
		return "just now"
	case seconds < 3600:
		m := seconds / 60
		return fmt.Sprintf("%dm ago", m)
	case seconds < 86400:
		h := seconds / 3600
		return fmt.Sprintf("%dh ago", h)
	case seconds < 604800:
		days := seconds / 86400
		return fmt.Sprintf("%dd ago", days)
	case seconds < 2592000:
		w := seconds / 604800
		return fmt.Sprintf("%dw ago", w)
	case seconds < 31536000:
		mo := seconds / 2592000
		return fmt.Sprintf("%dmo ago", mo)
	default:
		y := seconds / 31536000
		return fmt.Sprintf("%dy ago", y)
	}
}

// ContentHeight calculates the available height for the main content area.
func ContentHeight(termHeight int) int {
	h := termHeight - ChromeHeight
	if h < 1 {
		h = 1
	}
	return h
}

// TooSmall returns a message to display when the terminal is too small.
func TooSmall(width, height int) string {
	msg := "Terminal too small.\nResize to at least 40×10."
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center,
		lipgloss.NewStyle().Foreground(Subtle).Render(msg),
	)
}
