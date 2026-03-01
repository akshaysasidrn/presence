package main

import (
	"crypto/sha256"
	_ "embed"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

//go:embed quotes.json
var embeddedQuotesJSON []byte

type quote struct {
	Text   string `json:"text"`
	Author string `json:"author"`
}

// newQuote creates a quote with normalized text (smart punctuation replaced
// with ASCII equivalents so the text is typeable on a standard keyboard).
func newQuote(text, author string) quote {
	return quote{Text: normalizeText(text), Author: author}
}

func loadEmbeddedQuotes() ([]quote, error) {
	var quotes []quote
	if err := json.Unmarshal(embeddedQuotesJSON, &quotes); err != nil {
		return nil, fmt.Errorf("parsing embedded quotes: %w", err)
	}
	for i, q := range quotes {
		quotes[i] = newQuote(q.Text, q.Author)
	}
	return quotes, nil
}

func loadQuotesFromFile(path string) ([]quote, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading quotes file: %w", err)
	}
	var quotes []quote
	if err := json.Unmarshal(data, &quotes); err != nil {
		return nil, fmt.Errorf("parsing quotes file: %w", err)
	}
	if len(quotes) == 0 {
		return nil, fmt.Errorf("quotes file is empty")
	}
	for i, q := range quotes {
		quotes[i] = newQuote(q.Text, q.Author)
	}
	return quotes, nil
}

// normalizeText replaces smart/curly punctuation with ASCII equivalents
// so that quotes are typeable on a standard keyboard.
func normalizeText(s string) string {
	r := strings.NewReplacer(
		"\u2018", "'", // left single quote
		"\u2019", "'", // right single quote / apostrophe
		"\u201C", "\"", // left double quote
		"\u201D", "\"", // right double quote
		"\u2013", "-", // en dash
		"\u2014", "--", // em dash
		"\u2026", "...", // ellipsis
	)
	return r.Replace(s)
}

// dailyQuote picks a deterministic quote based on today's date.
func dailyQuote(quotes []quote) quote {
	today := time.Now().Format("2006-01-02")
	h := sha256.Sum256([]byte(today))
	idx := binary.BigEndian.Uint64(h[:8]) % uint64(len(quotes))
	return quotes[idx]
}

// randomQuote picks a random quote.
func randomQuote(quotes []quote) quote {
	return quotes[rand.Intn(len(quotes))]
}

// apiResponse matches the Stoic Quote API response shape.
// Usage: presence --api https://stoic.tekloon.net/stoic-quote
type apiResponse struct {
	Data struct {
		Author string `json:"author"`
		Quote  string `json:"quote"`
	} `json:"data"`
}

// fetchFromAPI fetches a quote from the given API endpoint.
// Returns (quote, false) on any error — caller should fall back to embedded quotes.
func fetchFromAPI(url string) (quote, bool) {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return quote{}, false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return quote{}, false
	}

	var result apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return quote{}, false
	}

	if result.Data.Quote == "" {
		return quote{}, false
	}

	return newQuote(result.Data.Quote, result.Data.Author), true
}
