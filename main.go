package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	tele "gopkg.in/telebot.v3"
)

var (
	botToken           = flag.String("token", os.Getenv("TELEGRAM_BOT_TOKEN"), "Telegram bot API token")
	botUrl             = flag.String("url", os.Getenv("TELEGRAM_BOT_URL"), "[Optional] Custom Telegram bot API URL")
	suppressTimestamps = flag.Bool("suppressTimestamps", false, "Specify this flag to omit timestamps in logs")
)

func main() {
	flag.Parse()

	if *suppressTimestamps {
		log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
	}

	if *botToken == "" {
		fmt.Println("Please provide a bot token with command line option '-token' or environment variable 'TELEGRAM_BOT_TOKEN'.")
		flag.Usage()
		os.Exit(1)
	}

	ctx := context.Background()
	logger := slog.Default()

	b, err := tele.NewBot(tele.Settings{
		URL:    *botUrl,
		Token:  *botToken,
		Poller: &tele.LongPoller{},
	})
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "Failed to create Telegram bot",
			slog.String("token", *botToken),
			slog.String("url", *botUrl),
			slog.Any("error", err),
		)
		os.Exit(1)
	}

	logger.LogAttrs(ctx, slog.LevelInfo, "Started Telegram bot",
		slog.String("username", b.Me.Username),
		slog.Int64("id", b.Me.ID),
	)

	b.Handle(tele.OnQuery, HandleInlineQuery)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		logger.LogAttrs(ctx, slog.LevelInfo, "Received exit signal", slog.Any("signal", sig))
		b.Stop()
	}()

	b.Start()
}
