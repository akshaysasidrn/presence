package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const version = "0.1.0"

var (
	correctStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#FAFAFA"})
	wrongStyle   = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#CC3333", Dark: "#FF4444"}).Background(lipgloss.AdaptiveColor{Light: "#FFD9D9", Dark: "#442222"})
	cursorStyle  = lipgloss.NewStyle().Underline(true).Foreground(lipgloss.AdaptiveColor{Light: "#666666", Dark: "#888888"})
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#AAAAAA", Dark: "#555555"})
	faintStyle   = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#CCCCCC", Dark: "#333333"})
)

var brailleStages = [3][]rune{
	[]rune("⣿⣾⣽⣻⣷⣯⣟⡿⢿"),
	[]rune("⠂⠃⠄⠅⠈⠐⠠⡀⢀"),
	[]rune("⠁⠂⠄⡀⠀"),
}

type (
	dissolveTickMsg struct{}
	doneMsg         struct{}
)

type keystroke struct {
	char    rune
	correct bool
}

type model struct {
	quote         []rune
	author        string
	typed         []keystroke
	done          bool
	width         int
	fleeting      bool
	dissolveFrame int
	dissolveRank  []int
}

func initialModel(q quote, fleeting bool) model {
	return model{
		quote:    []rune(q.Text),
		author:   q.Author,
		width:    80,
		fleeting: fleeting,
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

		case tea.KeyEnter:
			pos := len(m.typed)
			if pos < len(m.quote) && m.quote[pos] == '\n' {
				m.typed = append(m.typed, keystroke{'\n', true})
			}

		case tea.KeySpace:
			pos := len(m.typed)
			if pos < len(m.quote) {
				correct := m.quote[pos] == ' '
				m.typed = append(m.typed, keystroke{' ', correct})
			}

		case tea.KeyRunes:
			pos := len(m.typed)
			if pos < len(m.quote) {
				ch := msg.Runes[0]
				correct := ch == m.quote[pos]
				m.typed = append(m.typed, keystroke{ch, correct})
			}
		}

		// Check completion
		if len(m.typed) >= len(m.quote) {
			m.done = true
			if m.fleeting {
				attrRunes := []rune("— " + m.author)
				totalLen := len(m.quote) + len(attrRunes)
				perm := rand.Perm(totalLen)
				m.dissolveRank = make([]int, totalLen)
				for i, v := range perm {
					m.dissolveRank[v] = i
				}
				m.dissolveFrame = 1
				return m, tea.Tick(300*time.Millisecond, func(t time.Time) tea.Msg {
					return dissolveTickMsg{}
				})
			}
			return m, tea.Tick(800*time.Millisecond, func(t time.Time) tea.Msg {
				return doneMsg{}
			})
		}
		return m, nil

	case dissolveTickMsg:
		m.dissolveFrame++
		if m.dissolveFrame > 7 {
			return m, tea.Quit
		}
		return m, tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
			return dissolveTickMsg{}
		})

	case doneMsg:
		return m, tea.Quit
	}

	return m, nil
}

func (m model) View() string {
	if m.width == 0 {
		return ""
	}

	if m.dissolveFrame > 0 {
		return m.viewDissolve()
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
	pad := strings.Repeat(" ", padding)
	wrapped := strings.ReplaceAll(softWrap(styled.String(), m.quote, maxWidth), "\n", "\n"+pad)

	attrText := "— " + m.author
	attribution := dimStyle.Render(attrText)
	if m.done {
		attribution = correctStyle.Render(attrText)
	}

	return fmt.Sprintf("\n%s%s\n%s%s\n\n", pad, wrapped, pad, attribution)
}

func (m model) viewDissolve() string {
	padding := 4
	maxWidth := m.width - padding*2
	if maxWidth < 20 {
		maxWidth = 20
	}

	attrRunes := []rune("— " + m.author)
	totalLen := len(m.quote) + len(attrRunes)
	groupSize := totalLen / 3
	if groupSize < 1 {
		groupSize = 1
	}

	// Build styled quote
	var styled strings.Builder
	for i, r := range m.quote {
		styled.WriteString(m.dissolveChar(i, groupSize, r))
	}

	pad := strings.Repeat(" ", padding)
	wrapped := strings.ReplaceAll(softWrap(styled.String(), m.quote, maxWidth), "\n", "\n"+pad)

	// Build styled attribution
	var attrStyled strings.Builder
	for i, r := range attrRunes {
		attrStyled.WriteString(m.dissolveChar(len(m.quote)+i, groupSize, r))
	}

	return fmt.Sprintf("\n%s%s\n%s%s\n\n", pad, wrapped, pad, attrStyled.String())
}

func (m model) dissolveChar(idx, groupSize int, r rune) string {
	if r == ' ' || r == '\n' {
		return string(r)
	}

	rank := m.dissolveRank[idx]
	group := rank / groupSize
	if group > 2 {
		group = 2
	}
	startFrame := group + 1
	elapsed := m.dissolveFrame - startFrame + 1

	if elapsed <= 0 {
		return correctStyle.Render(string(r))
	}

	rng := rand.New(rand.NewSource(int64(idx*997 + m.dissolveFrame*31)))

	switch elapsed {
	case 1:
		pool := brailleStages[0]
		return dimStyle.Render(string(pool[rng.Intn(len(pool))]))
	case 2:
		pool := brailleStages[1]
		return dimStyle.Render(string(pool[rng.Intn(len(pool))]))
	case 3:
		pool := brailleStages[2]
		return faintStyle.Render(string(pool[rng.Intn(len(pool))]))
	default:
		return " "
	}
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
	dailyFlag := flag.Bool("daily", false, "pick the same quote for the whole day")
	fleetingFlag := flag.Bool("fleeting", false, "dissolve the quote into dust after completion")
	quotesFlag := flag.String("quotes", "", "path to a custom quotes JSON file")
	apiFlag := flag.String("api", "", "fetch a quote from an API endpoint (pass URL)")
	versionFlag := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("presence %s\n", version)
		return
	}

	// Select quote: --api > --quotes/default, then --daily or random
	var q quote
	if *apiFlag != "" {
		if fetched, ok := fetchFromAPI(*apiFlag); ok {
			q = fetched
		} else {
			// Fall back to embedded
			quotes, _ := loadEmbeddedQuotes()
			q = randomQuote(quotes)
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
		if *dailyFlag {
			q = dailyQuote(quotes)
		} else {
			q = randomQuote(quotes)
		}
	}

	p := tea.NewProgram(initialModel(q, *fleetingFlag))
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
