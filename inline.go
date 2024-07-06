package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"strings"
	"unicode"

	tele "gopkg.in/telebot.v3"
)

// NewHandleInlineQueryFunc returns a [tele.HandlerFunc] that handles inline queries.
func NewHandleInlineQueryFunc(ctx context.Context, logger *slog.Logger) tele.HandlerFunc {
	return func(c tele.Context) error {
		sender := c.Sender()
		text := c.Data()
		logger.LogAttrs(ctx, slog.LevelDebug, "Received inline query",
			slog.Int64("userID", sender.ID),
			slog.String("userFirstName", sender.FirstName),
			slog.String("username", sender.Username),
			slog.String("text", text),
		)

		results := tele.Results{
			&tele.ArticleResult{
				ResultBase: tele.ResultBase{
					ID: "addSpaces",
				},
				Title:       "🌌 I need some space!",
				Text:        addSpaces(text),
				Description: "Add extra spaces between each character in the message.",
			},
			&tele.ArticleResult{
				ResultBase: tele.ResultBase{
					ID: "randomizeCase",
				},
				Title:       "🦘 Jumpy Letters",
				Text:        randomizeCase(text),
				Description: "Randomly change letter case in the message.",
			},
			&tele.ArticleResult{
				ResultBase: tele.ResultBase{
					ID: "createTypos",
				},
				Title:       "✏️ feat: add typo",
				Text:        createTypos(text, 1),
				Description: "Randomly change the order of characters in the message.",
			},
			&tele.ArticleResult{
				ResultBase: tele.ResultBase{
					ID: "scrambleLetters",
				},
				Title:       "✍️ Scramble Letters",
				Text:        createTypos(text, 10+rand.IntN(10)),
				Description: "Recursively add typos.",
			},
			&tele.ArticleResult{
				ResultBase: tele.ResultBase{
					ID: "generateMe",
				},
				Title:       "🤳 What the hell am I doing?",
				Text:        generateMe(sender, text),
				Description: "Tell everyone what you're doing (/me).",
			},
			&tele.ArticleResult{
				ResultBase: tele.ResultBase{
					ID: "repeat",
				},
				Title:       "🔂 Can you repeat what I just said?",
				Text:        repeat(text),
				Description: "Repeat the message three times.",
			},
			&tele.ArticleResult{
				ResultBase: tele.ResultBase{
					ID: "reverse",
				},
				Title:       "🔀 上海自来水",
				Text:        reverse(text),
				Description: "Reverse the order of characters in the message.",
			},
			&tele.ArticleResult{
				ResultBase: tele.ResultBase{
					ID: "mirror",
				},
				Title:       "🪞 上海自来水来自海上",
				Text:        mirror(text),
				Description: "Mirror the message in reverse order.",
			},
			&tele.ArticleResult{
				ResultBase: tele.ResultBase{
					ID: "comboSpacesRepeat",
				},
				Title:       "🛠️ Combo: Spaces + Repeat",
				Text:        repeat(addSpaces(text)),
				Description: "Add extra spaces between each character. Then repeat the message three times.",
			},
			&tele.ArticleResult{
				ResultBase: tele.ResultBase{
					ID: "comboRandomcaseSpaces",
				},
				Title:       "🛠️ Combo: Random Case + Spaces",
				Text:        addSpaces(randomizeCase(text)),
				Description: "Randomly change letter case. Then add extra spaces between each character.",
			},
		}

		return c.Answer(&tele.QueryResponse{
			Results:   results,
			CacheTime: 1,
		})
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

	for range times {
		// Swap runes[pos] and runes[pos+1]
		pos := rand.IntN(len(runes) - 1)
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

	var (
		ri        uint64
		remaining int
	)

	for i, r := range runes {
		if uint32(r|0x20-'a') > 'z'-'a' {
			continue
		}

		if remaining == 0 {
			ri, remaining = rand.Uint64(), 64
		}

		if ri&1 == 1 {
			runes[i] = r ^ 0x20
		}

		ri >>= 1
		remaining--
	}

	return string(runes)
}

// reverse reverses the order of runes in s.
func reverse(s string) string {
	if s == "" {
		s = "上海自来水"
	}

	runes := []rune(s)

	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}

// mirror mirrors the message in reverse order.
func mirror(s string) string {
	if s == "" {
		s = "上海自来水"
	}

	originalRunes := []rune(s)
	runes := make([]rune, len(originalRunes)*2-1)
	copy(runes, originalRunes)

	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[j] = runes[i]
	}

	return string(runes)
}
