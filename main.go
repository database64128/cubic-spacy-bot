package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	tb "gopkg.in/tucnak/telebot.v2"
)

func main() {
	var botToken string
	var botUrl string
	var suppressTimestamps bool

	flag.StringVar(&botToken, "token", os.Getenv("TELEGRAM_BOT_TOKEN"), "Telegram bot API token")
	flag.StringVar(&botUrl, "url", os.Getenv("TELEGRAM_BOT_URL"), "[Optional] Custom Telegram bot API URL")
	flag.BoolVar(&suppressTimestamps, "suppressTimestamps", false, "Specify this flag to omit timestamps in logs")

	flag.Parse()

	if suppressTimestamps {
		log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
	}

	if botToken == "" {
		log.Fatal("Please provide a bot token with command line option '-token' or environment variable 'TELEGRAM_BOT_TOKEN'.")
	}

	b, err := tb.NewBot(tb.Settings{
		URL:    botUrl,
		Token:  botToken,
		Poller: &tb.LongPoller{},
	})

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Started Telegram bot: @%s (%d)", b.Me.Username, b.Me.ID)

	b.Handle(tb.OnQuery, func(q *tb.Query) {
		HandleInlineQuery(b, q)
	})

	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.Printf("Received %s, shutting down...", sig.String())
		b.Stop()
		os.Exit(0)
	}()

	b.Start()
}
