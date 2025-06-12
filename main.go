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

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/lmittmann/tint"
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

	opts := make([]bot.Option, 0, 4)

	if botURL != "" {
		opts = append(opts, bot.WithServerURL(botURL))
	}

	opts = append(opts,
		bot.WithSkipGetMe(),
		bot.WithDefaultHandler(NewInlineQueryHandler(logger)),
		bot.WithErrorsHandler(func(err error) {
			logger.LogAttrs(ctx, slog.LevelWarn, "Failed to handle update",
				tint.Err(err),
			)
		}),
	)

	b, err := bot.New(botToken, opts...)
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "Failed to create Telegram bot",
			slog.String("token", botToken),
			slog.String("url", botURL),
			tint.Err(err),
		)
		os.Exit(1)
	}

	var me *models.User

	for {
		me, err = b.GetMe(ctx)
		if err != nil {
			logger.LogAttrs(ctx, slog.LevelError, "Failed to get bot info, retrying in 30 seconds",
				slog.String("token", botToken),
				slog.String("url", botURL),
				tint.Err(err),
			)
			time.Sleep(30 * time.Second)
			continue
		}
		break
	}

	logger.LogAttrs(ctx, slog.LevelInfo, "Started Telegram bot",
		slog.String("username", me.Username),
		slog.Int64("id", me.ID),
	)

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-ctx.Done()
		logger.LogAttrs(ctx, slog.LevelInfo, "Received exit signal")
		stop()
	}()

	b.Start(ctx)
}
