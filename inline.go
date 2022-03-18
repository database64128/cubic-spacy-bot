package main

import (
	"fmt"
	"math/rand"
	"strings"
	"unicode"

	tele "gopkg.in/telebot.v3"
)

// HandleInlineQuery handles an inline query.
func HandleInlineQuery(c tele.Context) error {
	text := c.Text()
	sender := c.Sender()
	results := make(tele.Results, 8)

	results[0] = &tele.ArticleResult{
		Title:       "🌌 I need some space!",
		Description: "Add extra spaces between each character in the message.",
		Text:        addSpaces(text),
	}

	results[1] = &tele.ArticleResult{
		Title:       "🦘 Jumpy Letters",
		Description: "Randomly change letter case in the message.",
		Text:        randomizeCase(text),
	}

	results[2] = &tele.ArticleResult{
		Title:       "✏️ feat: add typo",
		Description: "Randomly change the order of characters in the message.",
		Text:        createTypos(text, 1),
	}

	results[3] = &tele.ArticleResult{
		Title:       "✍️ Scramble Letters",
		Description: "Recursively add typos.",
		Text:        createTypos(text, 10+rand.Intn(10)),
	}

	results[4] = &tele.ArticleResult{
		Title:       "🤳 What the hell am I doing?",
		Description: "Tell everyone what you're doing (/me).",
		Text:        generateMe(sender, text),
	}

	results[5] = &tele.ArticleResult{
		Title:       "🔂 Can you repeat what I just said?",
		Description: "Repeat the message three times.",
		Text:        repeat(text),
	}

	results[6] = &tele.ArticleResult{
		Title:       "🛠️ Combo: Spaces + Repeat",
		Description: "Add extra spaces between each character. Then repeat the message three times.",
		Text:        repeat(addSpaces(text)),
	}

	results[7] = &tele.ArticleResult{
		Title:       "🛠️ Combo: Random Case + Spaces",
		Description: "Randomly change letter case. Then add extra spaces between each character.",
		Text:        addSpaces(randomizeCase(text)),
	}

	results[0].SetResultID("addSpaces")
	results[1].SetResultID("randomizeCase")
	results[2].SetResultID("createTypos")
	results[3].SetResultID("scrambleLetters")
	results[4].SetResultID("generateMe")
	results[5].SetResultID("repeat")
	results[6].SetResultID("comboSpacesRepeat")
	results[7].SetResultID("comboRandomcaseSpaces")

	return c.Answer(&tele.QueryResponse{
		Results:   results,
		CacheTime: 1,
	})
}

// addSpaces adds one space between ASCII characters, two spaces between non-ASCII characters.
func addSpaces(s string) string {
	if s == "" {
		s = "🌌 I need some space!"
	}

	var sb strings.Builder

	for _, r := range s {
		switch {
		case r > unicode.MaxASCII:
			sb.WriteByte(' ')
			sb.WriteRune(r)
			sb.WriteByte(' ')
		default:
			sb.WriteRune(r)
			sb.WriteByte(' ')
		}
	}

	return strings.TrimSpace(sb.String())
}

// createTypos creates typos in the input message by randomly changing the order of characters.
func createTypos(s string, rounds int) string {
	if s == "" {
		s = "✏️ feat: add typo"
	}

	runes := []rune(s)

	if len(runes) < 2 {
		return s
	}

	times := (1 + len(runes)/20) * rounds

	for i := 0; i < times; i++ {
		// Swap runes[pos] and runes[pos+1]
		pos := rand.Intn(len(runes) - 1)
		runes[pos], runes[pos+1] = runes[pos+1], runes[pos]
	}

	return string(runes)
}

// generateMe generates a '/me' message.
func generateMe(from *tele.User, s string) string {
	if s == "" {
		s = "doesn't know what to say. 🤐"
	}

	return fmt.Sprintf("* %s %s", from.FirstName, s)
}

// repeat repeats the message three times.
func repeat(s string) string {
	if s == "" {
		s = "I repeat!"
	}

	return fmt.Sprintf("%s\n%s\n%s", s, s, s)
}

// randomizeCase randomizes letter case in the message by randomly use .ToLower or .ToUpper on letters.
func randomizeCase(s string) string {
	if s == "" {
		s = "The quick brown fox jumps over the lazy dog."
	}

	runes := []rune(s)

	for i := range runes {
		isLower := 'a' <= runes[i] && runes[i] <= 'z'
		isUpper := 'A' <= runes[i] && runes[i] <= 'Z'

		if !isLower && !isUpper {
			continue
		}

		switch rand.Intn(2) {
		case 0: // ToUpper
			if isLower {
				runes[i] -= 'a' - 'A'
			}
		case 1: // ToLower
			if isUpper {
				runes[i] += 'a' - 'A'
			}
		}
	}

	return string(runes)
}
