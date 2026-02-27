package main

import (
	"fmt"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	quote  = "The impediment to action advances action. What stands in the way becomes the way."
	author = "Marcus Aurelius"
)

var (
	correctStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA"))
	wrongStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444"))
	cursorStyle  = lipgloss.NewStyle().Underline(true).Foreground(lipgloss.Color("#888888"))
	ghostStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	authorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
)

type doneMsg struct{}

type model struct {
	quote   string
	author  string
	typed   []rune
	correct []bool
	done    bool
	width   int
}

func initialModel() model {
	return model{
		quote:  quote,
		author: author,
		width:  80,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil

	case tea.KeyMsg:
		if m.done {
			return m, nil
		}

		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyBackspace:
			if len(m.typed) > 0 {
				m.typed = m.typed[:len(m.typed)-1]
				m.correct = m.correct[:len(m.correct)-1]
			}
			return m, nil

		case tea.KeyEnter:
			runes := []rune(m.quote)
			pos := len(m.typed)
			if pos < len(runes) && runes[pos] == '\n' {
				m.typed = append(m.typed, '\n')
				m.correct = append(m.correct, true)
			}

		case tea.KeySpace:
			runes := []rune(m.quote)
			pos := len(m.typed)
			if pos < len(runes) {
				m.typed = append(m.typed, ' ')
				m.correct = append(m.correct, runes[pos] == ' ')
			}

		case tea.KeyRunes:
			runes := []rune(m.quote)
			pos := len(m.typed)
			if pos < len(runes) {
				ch := msg.Runes[0]
				m.typed = append(m.typed, ch)
				m.correct = append(m.correct, ch == runes[pos])
			}
		}

		// Check completion
		quoteRunes := []rune(m.quote)
		if len(m.typed) >= len(quoteRunes) {
			m.done = true
			return m, tea.Tick(800*time.Millisecond, func(t time.Time) tea.Msg {
				return doneMsg{}
			})
		}
		return m, nil

	case doneMsg:
		return m, tea.Quit
	}

	return m, nil
}

func (m model) View() string {
	if m.width == 0 {
		return ""
	}

	padding := 4
	maxWidth := m.width - padding*2
	if maxWidth < 20 {
		maxWidth = 20
	}

	quoteRunes := []rune(m.quote)
	var styled strings.Builder

	for i, r := range quoteRunes {
		ch := string(r)
		if i < len(m.typed) {
			if m.correct[i] {
				styled.WriteString(correctStyle.Render(ch))
			} else {
				styled.WriteString(wrongStyle.Render(ch))
			}
		} else if i == len(m.typed) {
			styled.WriteString(cursorStyle.Render(ch))
		} else {
			styled.WriteString(ghostStyle.Render(ch))
		}
	}

	// Word-wrap the styled text
	wrapped := softWrap(styled.String(), quoteRunes, maxWidth)

	pad := strings.Repeat(" ", padding)
	attribution := authorStyle.Render("— " + m.author)

	return fmt.Sprintf("\n%s%s\n%s%s\n\n", pad, wrapped, pad, attribution)
}

// softWrap wraps styled text at word boundaries based on the raw rune widths.
func softWrap(styled string, raw []rune, maxWidth int) string {
	segments := splitStyledSegments(styled, len(raw))

	// Find indices where we should break (replace space with newline).
	breakAt := make(map[int]bool)
	col := 0
	lastSpace := -1
	for i, r := range raw {
		if r == ' ' {
			lastSpace = i
		}
		col++
		if col > maxWidth && lastSpace >= 0 {
			breakAt[lastSpace] = true
			col = i - lastSpace
			lastSpace = -1
		}
	}

	var result strings.Builder
	for i := range raw {
		if breakAt[i] {
			result.WriteString("\n")
		} else {
			result.WriteString(segments[i])
		}
	}
	return result.String()
}

// splitStyledSegments extracts one styled segment per raw rune from the
// ANSI-styled string. Each segment includes the ANSI escape codes wrapping
// that character.
func splitStyledSegments(styled string, count int) []string {
	segments := make([]string, 0, count)
	i := 0
	bytes := []byte(styled)

	for len(segments) < count && i < len(bytes) {
		start := i
		// Consume any leading ANSI escape sequences
		for i < len(bytes) && bytes[i] == '\x1b' {
			for i < len(bytes) && bytes[i] != 'm' {
				i++
			}
			if i < len(bytes) {
				i++ // skip 'm'
			}
		}
		// Consume one UTF-8 character
		if i < len(bytes) {
			_, size := utf8.DecodeRune(bytes[i:])
			i += size
		}
		// Consume any trailing ANSI escape sequences (reset codes)
		for i < len(bytes) && bytes[i] == '\x1b' {
			for i < len(bytes) && bytes[i] != 'm' {
				i++
			}
			if i < len(bytes) {
				i++ // skip 'm'
			}
		}
		segments = append(segments, string(bytes[start:i]))
	}

	return segments
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
