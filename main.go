package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"runtime/debug"
	"strconv"
	"syscall"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/lmittmann/tint"
)

var (
	version    bool
	logNoColor bool
	logNoTime  bool
	logLevel   slog.Level
	botToken   string
	botURL     string

	botWebhookListenNetwork string
	botWebhookListenAddress string
	botWebhookListenOwner   string
	botWebhookListenGroup   string
	botWebhookListenMode    string

	botWebhookSecretToken string
	botWebhookURL         string
)

func init() {
	flag.BoolVar(&version, "version", false, "Print version and exit")
	flag.BoolVar(&logNoColor, "logNoColor", false, "Disable colors in log output")
	flag.BoolVar(&logNoTime, "logNoTime", false, "Disable timestamps in log output")
	flag.TextVar(&logLevel, "logLevel", slog.LevelInfo, "Log level")
	flag.StringVar(&botToken, "token", os.Getenv("TELEGRAM_BOT_TOKEN"), "Telegram bot API token")
	flag.StringVar(&botURL, "url", os.Getenv("TELEGRAM_BOT_URL"), "[Optional] Custom Telegram bot API URL")

	flag.StringVar(&botWebhookListenNetwork, "webhookListenNetwork", os.Getenv("TELEGRAM_BOT_WEBHOOK_LISTEN_NETWORK"), "Network for webhook listener (e.g., tcp, unix)")
	flag.StringVar(&botWebhookListenAddress, "webhookListenAddress", os.Getenv("TELEGRAM_BOT_WEBHOOK_LISTEN_ADDRESS"), "Address for webhook listener (e.g., :8080, /run/cubic-spacy-bot.sock)")
	flag.StringVar(&botWebhookListenOwner, "webhookListenOwner", os.Getenv("TELEGRAM_BOT_WEBHOOK_LISTEN_OWNER"), "Owner for webhook unix domain socket (e.g., http)")
	flag.StringVar(&botWebhookListenGroup, "webhookListenGroup", os.Getenv("TELEGRAM_BOT_WEBHOOK_LISTEN_GROUP"), "Group for webhook unix domain socket (e.g., http)")
	flag.StringVar(&botWebhookListenMode, "webhookListenMode", os.Getenv("TELEGRAM_BOT_WEBHOOK_LISTEN_MODE"), "File mode for webhook unix domain socket (e.g., 0660)")

	flag.StringVar(&botWebhookSecretToken, "webhookSecretToken", os.Getenv("TELEGRAM_BOT_WEBHOOK_SECRET_TOKEN"), "[Optional] Secret token for webhook authentication")
	flag.StringVar(&botWebhookURL, "webhookURL", os.Getenv("TELEGRAM_BOT_WEBHOOK_URL"), "Webhook URL to set for the bot (e.g., https://example.com/cubic-spacy-bot)")
}

func main() {
	flag.Parse()

	if version {
		if info, ok := debug.ReadBuildInfo(); ok {
			os.Stdout.WriteString(info.String())
		}
		return
	}

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

	opts := make([]bot.Option, 0, 6)

	if botURL != "" {
		opts = append(opts, bot.WithServerURL(botURL))
	}

	opts = append(opts,
		bot.WithSkipGetMe(),
		bot.WithWebhookSecretToken(botWebhookSecretToken),
		bot.WithDefaultHandler(NewInlineQueryHandler(logger)),
		bot.WithErrorsHandler(func(err error) {
			logger.LogAttrs(ctx, slog.LevelWarn, "Failed to handle update",
				tint.Err(err),
			)
		}),
		bot.WithAllowedUpdates(bot.AllowedUpdates{models.AllowedUpdateInlineQuery}),
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

	for {
		if _, err = b.SetWebhook(ctx, &bot.SetWebhookParams{
			URL:            botWebhookURL,
			AllowedUpdates: []string{models.AllowedUpdateInlineQuery},
			SecretToken:    botWebhookSecretToken,
		}); err != nil {
			logger.LogAttrs(ctx, slog.LevelError, "Failed to set webhook, retrying in 30 seconds",
				slog.String("webhookURL", botWebhookURL),
				slog.String("webhookSecretToken", botWebhookSecretToken),
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

	if botWebhookURL != "" {
		runWebhookServer(ctx, logger, b)
	} else {
		b.Start(ctx)
	}
}

func runWebhookServer(ctx context.Context, logger *slog.Logger, b *bot.Bot) {
	var lc net.ListenConfig
	ln, err := lc.Listen(ctx, botWebhookListenNetwork, botWebhookListenAddress)
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "Failed to start webhook listener",
			slog.String("network", botWebhookListenNetwork),
			slog.String("address", botWebhookListenAddress),
			tint.Err(err),
		)
		os.Exit(1)
	}
	listenAddress := ln.Addr().String()

	if botWebhookListenNetwork == "unix" {
		if botWebhookListenOwner != "" || botWebhookListenGroup != "" {
			uid, gid := -1, -1

			if botWebhookListenOwner != "" {
				var uidString string

				owner, err := user.Lookup(botWebhookListenOwner)
				if err != nil {
					var e user.UnknownUserError
					if errors.As(err, &e) {
						uidString = botWebhookListenOwner
					} else {
						logger.LogAttrs(ctx, slog.LevelError, "Failed to lookup user for webhook socket owner",
							slog.String("owner", botWebhookListenOwner),
							tint.Err(err),
						)
						os.Exit(1)
					}
				} else {
					uidString = owner.Uid
				}

				uid, err = strconv.Atoi(uidString)
				if err != nil {
					logger.LogAttrs(ctx, slog.LevelError, "Invalid username or uid for webhook socket owner",
						slog.String("owner", botWebhookListenOwner),
						slog.String("uid", uidString),
						tint.Err(err),
					)
					os.Exit(1)
				}
			}

			if botWebhookListenGroup != "" {
				var gidString string

				group, err := user.LookupGroup(botWebhookListenGroup)
				if err != nil {
					var e user.UnknownGroupError
					if errors.As(err, &e) {
						gidString = botWebhookListenGroup
					} else {
						logger.LogAttrs(ctx, slog.LevelError, "Failed to lookup group for webhook socket group",
							slog.String("group", botWebhookListenGroup),
							tint.Err(err),
						)
						os.Exit(1)
					}
				} else {
					gidString = group.Gid
				}

				gid, err = strconv.Atoi(gidString)
				if err != nil {
					logger.LogAttrs(ctx, slog.LevelError, "Invalid group name or gid for webhook socket group",
						slog.String("group", botWebhookListenGroup),
						slog.String("gid", gidString),
						tint.Err(err),
					)
					os.Exit(1)
				}
			}

			if err := os.Chown(listenAddress, uid, gid); err != nil {
				logger.LogAttrs(ctx, slog.LevelError, "Failed to set owner/group for webhook socket",
					slog.String("address", listenAddress),
					slog.String("owner", botWebhookListenOwner),
					slog.String("group", botWebhookListenGroup),
					slog.Int("uid", uid),
					slog.Int("gid", gid),
					tint.Err(err),
				)
				os.Exit(1)
			}

			logger.LogAttrs(ctx, slog.LevelDebug, "Set owner/group for webhook socket",
				slog.String("address", listenAddress),
				slog.String("owner", botWebhookListenOwner),
				slog.String("group", botWebhookListenGroup),
				slog.Int("uid", uid),
				slog.Int("gid", gid),
			)
		}

		if botWebhookListenMode != "" {
			mode, err := strconv.ParseUint(botWebhookListenMode, 8, 32)
			if err != nil {
				logger.LogAttrs(ctx, slog.LevelError, "Invalid file mode for webhook socket",
					slog.String("mode", botWebhookListenMode),
					tint.Err(err),
				)
				os.Exit(1)
			}
			fileMode := fs.FileMode(mode)

			if err := os.Chmod(listenAddress, fileMode); err != nil {
				logger.LogAttrs(ctx, slog.LevelError, "Failed to set file mode for webhook socket",
					slog.String("address", listenAddress),
					slog.Any("mode", fileMode),
					tint.Err(err),
				)
				os.Exit(1)
			}

			logger.LogAttrs(ctx, slog.LevelDebug, "Set file mode for webhook socket",
				slog.String("address", listenAddress),
				slog.Any("mode", fileMode),
			)
		}
	}

	server := http.Server{
		Addr:     listenAddress,
		Handler:  b.WebhookHandler(),
		ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	go func() {
		if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
			logger.LogAttrs(ctx, slog.LevelError, "Failed to serve webhook", tint.Err(err))
			os.Exit(1)
		}
	}()

	logger.LogAttrs(ctx, slog.LevelInfo, "Started webhook server", slog.String("listenAddress", listenAddress))

	b.StartWebhook(ctx)
	_ = server.Close()
}
