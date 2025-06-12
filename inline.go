package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"slices"
	"strings"
	"unicode"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// NewInlineQueryHandler returns a handler for inline queries.
func NewInlineQueryHandler(logger *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		inlineQuery := update.InlineQuery
		if inlineQuery == nil {
			return
		}
		sender := inlineQuery.From
		text := inlineQuery.Query
		logger.LogAttrs(ctx, slog.LevelDebug, "Received inline query",
			slog.Int64("userID", sender.ID),
			slog.String("userFirstName", sender.FirstName),
			slog.String("username", sender.Username),
			slog.String("text", text),
		)

		articles := []models.InlineQueryResultArticle{
			{
				ID:    "addSpaces",
				Title: "🌌 I need some space!",
				InputMessageContent: models.InputTextMessageContent{
					MessageText: addSpaces(text),
				},
				Description: "Add extra spaces between each character in the message.",
			},
			{
				ID:    "randomizeCase",
				Title: "🦘 Jumpy Letters",
				InputMessageContent: models.InputTextMessageContent{
					MessageText: randomizeCase(text),
				},
				Description: "Randomly change letter case in the message.",
			},
			{
				ID:    "createTypos",
				Title: "✏️ feat: add typo",
				InputMessageContent: models.InputTextMessageContent{
					MessageText: createTypos(text, 1),
				},
				Description: "Randomly change the order of characters in the message.",
			},
			{
				ID:    "scrambleLetters",
				Title: "✍️ Scramble Letters",
				InputMessageContent: models.InputTextMessageContent{
					MessageText: createTypos(text, 10+rand.IntN(10)),
				},
				Description: "Recursively add typos.",
			},
			{
				ID:    "generateMe",
				Title: "🤳 What the hell am I doing?",
				InputMessageContent: models.InputTextMessageContent{
					MessageText: generateMe(sender, text),
				},
				Description: "Tell everyone what you're doing (/me).",
			},
			{
				ID:    "repeat",
				Title: "🔂 Can you repeat what I just said?",
				InputMessageContent: models.InputTextMessageContent{
					MessageText: repeat(text),
				},
				Description: "Repeat the message three times.",
			},
			{
				ID:    "reverse",
				Title: "🔀 上海自来水",
				InputMessageContent: models.InputTextMessageContent{
					MessageText: reverse(text),
				},
				Description: "Reverse the order of characters in the message.",
			},
			{
				ID:    "mirror",
				Title: "🪞 上海自来水来自海上",
				InputMessageContent: models.InputTextMessageContent{
					MessageText: mirror(text),
				},
				Description: "Mirror the message in reverse order.",
			},
			{
				ID:    "comboSpacesRepeat",
				Title: "🛠️ Combo: Spaces + Repeat",
				InputMessageContent: models.InputTextMessageContent{
					MessageText: repeat(addSpaces(text)),
				},
				Description: "Add extra spaces between each character. Then repeat the message three times.",
			},
			{
				ID:    "comboRandomcaseSpaces",
				Title: "🛠️ Combo: Random Case + Spaces",
				InputMessageContent: models.InputTextMessageContent{
					MessageText: addSpaces(randomizeCase(text)),
				},
				Description: "Randomly change letter case. Then add extra spaces between each character.",
			},
		}

		results := make([]models.InlineQueryResult, len(articles))
		for i, article := range articles {
			results[i] = &article
		}

		b.AnswerInlineQuery(ctx, &bot.AnswerInlineQueryParams{
			InlineQueryID: update.InlineQuery.ID,
			Results:       results,
			CacheTime:     1,
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
func generateMe(from *models.User, s string) string {
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
	slices.Reverse(runes)
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
