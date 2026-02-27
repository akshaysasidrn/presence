package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const version = "0.2.0"

var (
	correctStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#FAFAFA"})
	wrongStyle   = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#CC3333", Dark: "#FF4444"})
	cursorStyle  = lipgloss.NewStyle().Underline(true).Foreground(lipgloss.AdaptiveColor{Light: "#666666", Dark: "#888888"})
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#AAAAAA", Dark: "#555555"})
)

type doneMsg struct{}

type keystroke struct {
	char    rune
	correct bool
}

type model struct {
	quote  []rune
	author string
	typed  []keystroke
	done   bool
	width  int
}

func initialModel(q quote) model {
	return model{
		quote:  []rune(q.Text),
		author: q.Author,
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
		case tea.KeyCtrlC, tea.KeyEsc, tea.KeyTab:
			return m, tea.Quit

		case tea.KeyBackspace:
			if len(m.typed) > 0 {
				m.typed = m.typed[:len(m.typed)-1]
			}
			return m, nil

		case tea.KeyEnter:
			pos := len(m.typed)
			if pos < len(m.quote) && m.quote[pos] == '\n' {
				m.typed = append(m.typed, keystroke{'\n', true})
			}

		case tea.KeySpace:
			pos := len(m.typed)
			if pos < len(m.quote) {
				m.typed = append(m.typed, keystroke{' ', m.quote[pos] == ' '})
			}

		case tea.KeyRunes:
			pos := len(m.typed)
			if pos < len(m.quote) {
				ch := msg.Runes[0]
				m.typed = append(m.typed, keystroke{ch, ch == m.quote[pos]})
			}
		}

		// Check completion
		if len(m.typed) >= len(m.quote) {
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

	var styled strings.Builder

	for i, r := range m.quote {
		ch := string(r)
		if i < len(m.typed) {
			if m.typed[i].correct {
				styled.WriteString(correctStyle.Render(ch))
			} else {
				styled.WriteString(wrongStyle.Render(ch))
			}
		} else if i == len(m.typed) {
			styled.WriteString(cursorStyle.Render(ch))
		} else {
			styled.WriteString(dimStyle.Render(ch))
		}
	}

	// Word-wrap the styled text
	wrapped := softWrap(styled.String(), m.quote, maxWidth)

	pad := strings.Repeat(" ", padding)
	attribution := dimStyle.Render("— " + m.author)

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

	skipANSI := func() {
		for i < len(bytes) && bytes[i] == '\x1b' {
			for i < len(bytes) && bytes[i] != 'm' {
				i++
			}
			if i < len(bytes) {
				i++ // skip 'm'
			}
		}
	}

	for len(segments) < count && i < len(bytes) {
		start := i
		skipANSI()
		// Consume one UTF-8 character
		if i < len(bytes) {
			_, size := utf8.DecodeRune(bytes[i:])
			i += size
		}
		skipANSI()
		segments = append(segments, string(bytes[start:i]))
	}

	return segments
}

func main() {
	randomFlag := flag.Bool("random", false, "pick a random quote")
	quotesFlag := flag.String("quotes", "", "path to a custom quotes JSON file")
	apiFlag := flag.Bool("api", false, "fetch a quote from the Stoic Quote API")
	versionFlag := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("presence %s\n", version)
		return
	}

	// Select quote: --quotes > --api > --random > daily
	var q quote
	if *apiFlag {
		if fetched, ok := fetchFromAPI(); ok {
			q = fetched
		} else {
			// Fall back to embedded
			quotes, _ := loadEmbeddedQuotes()
			q = dailyQuote(quotes)
		}
	} else {
		var quotes []quote
		var err error
		if *quotesFlag != "" {
			quotes, err = loadQuotesFromFile(*quotesFlag)
		} else {
			quotes, err = loadEmbeddedQuotes()
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if *randomFlag {
			q = randomQuote(quotes)
		} else {
			q = dailyQuote(quotes)
		}
	}

	p := tea.NewProgram(initialModel(q))
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
