package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"unicode"

	tb "gopkg.in/tucnak/telebot.v2"
)

// HandleInlineQuery handles an inline query.
func HandleInlineQuery(b *tb.Bot, q *tb.Query) {
	results := make(tb.Results, 8)

	results[0] = &tb.ArticleResult{
		Title:       "🌌 I need some space!",
		Description: "Add extra spaces between each character in the message.",
		Text:        addSpaces(q.Text),
	}

	results[1] = &tb.ArticleResult{
		Title:       "🦘 Jumpy Letters",
		Description: "Randomly change letter case in the message.",
		Text:        randomizeCase(q.Text),
	}

	results[2] = &tb.ArticleResult{
		Title:       "✏️ feat: add typo",
		Description: "Randomly change the order of characters in the message.",
		Text:        createTypos(q.Text, 1),
	}

	results[3] = &tb.ArticleResult{
		Title:       "✍️ Scramble Letters",
		Description: "Recursively add typos.",
		Text:        createTypos(q.Text, 10+rand.Intn(10)),
	}

	results[4] = &tb.ArticleResult{
		Title:       "🤳 What the hell am I doing?",
		Description: "Tell everyone what you're doing (/me).",
		Text:        generateMe(q.From, q.Text),
	}

	results[5] = &tb.ArticleResult{
		Title:       "🔂 Can you repeat what I just said?",
		Description: "Repeat the message three times.",
		Text:        repeat(q.Text),
	}

	results[6] = &tb.ArticleResult{
		Title:       "🛠️ Combo: Spaces + Repeat",
		Description: "Add extra spaces between each character. Then repeat the message three times.",
		Text:        repeat(addSpaces(q.Text)),
	}

	results[7] = &tb.ArticleResult{
		Title:       "🛠️ Combo: Random Case + Spaces",
		Description: "Randomly change letter case. Then add extra spaces between each character.",
		Text:        addSpaces(randomizeCase(q.Text)),
	}

	results[0].SetResultID("addSpaces")
	results[1].SetResultID("randomizeCase")
	results[2].SetResultID("createTypos")
	results[3].SetResultID("scrambleLetters")
	results[4].SetResultID("generateMe")
	results[5].SetResultID("repeat")
	results[6].SetResultID("comboSpacesRepeat")
	results[7].SetResultID("comboRandomcaseSpaces")

	err := b.Answer(q, &tb.QueryResponse{
		Results:   results,
		CacheTime: 1,
	})

	if err != nil {
		log.Println(err)
	}
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
func generateMe(from tb.User, s string) string {
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
