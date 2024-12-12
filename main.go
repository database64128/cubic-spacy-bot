package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lmittmann/tint"
	tele "gopkg.in/telebot.v3"
)

var (
	logNoColor bool
	logNoTime  bool
	logLevel   slog.Level
	botToken   string
	botURL     string
)

func init() {
	flag.BoolVar(&logNoColor, "logNoColor", false, "Disable colors in log output")
	flag.BoolVar(&logNoTime, "logNoTime", false, "Disable timestamps in log output")
	flag.TextVar(&logLevel, "logLevel", slog.LevelInfo, "Log level")
	flag.StringVar(&botToken, "token", os.Getenv("TELEGRAM_BOT_TOKEN"), "Telegram bot API token")
	flag.StringVar(&botURL, "url", os.Getenv("TELEGRAM_BOT_URL"), "[Optional] Custom Telegram bot API URL")
}

func main() {
	flag.Parse()

	if botToken == "" {
		fmt.Fprintln(os.Stderr, "Please provide a bot token with command line option '-token' or environment variable 'TELEGRAM_BOT_TOKEN'.")
		flag.Usage()
		os.Exit(1)
	}

	var replaceAttr func(groups []string, attr slog.Attr) slog.Attr
	if logNoTime {
		replaceAttr = func(groups []string, attr slog.Attr) slog.Attr {
			if len(groups) == 0 && attr.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return attr
		}
	}

	logger := slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:       logLevel,
		ReplaceAttr: replaceAttr,
		NoColor:     logNoColor,
	}))

	ctx := context.Background()

	var (
		b   *tele.Bot
		err error
	)

	for {
		b, err = tele.NewBot(tele.Settings{
			URL:   botURL,
			Token: botToken,
			OnError: func(err error, _ tele.Context) {
				logger.LogAttrs(ctx, slog.LevelWarn, "Failed to handle update",
					slog.Any("error", err),
				)
			},
		})
		if err != nil {
			logger.LogAttrs(ctx, slog.LevelError, "Failed to create Telegram bot, retrying in 30 seconds",
				slog.String("token", botToken),
				slog.String("url", botURL),
				slog.Any("error", err),
			)
			time.Sleep(30 * time.Second)
			continue
		}
		break
	}

	logger.LogAttrs(ctx, slog.LevelInfo, "Started Telegram bot",
		slog.String("username", b.Me.Username),
		slog.Int64("id", b.Me.ID),
	)

	b.Handle(tele.OnQuery, NewHandleInlineQueryFunc(ctx, logger))

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		logger.LogAttrs(ctx, slog.LevelInfo, "Received exit signal", slog.Any("signal", sig))
		signal.Stop(sigCh)
		b.Stop()
	}()

	b.Start()
}
