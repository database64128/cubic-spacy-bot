package main

import (
	"flag"
	"fmt"
	"log"
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

	b, err := tele.NewBot(tele.Settings{
		URL:    *botUrl,
		Token:  *botToken,
		Poller: &tele.LongPoller{},
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Started Telegram bot: @%s (%d)", b.Me.Username, b.Me.ID)

	b.Handle(tele.OnQuery, HandleInlineQuery)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		log.Printf("Received %s, stopping...", sig.String())
		b.Stop()
	}()

	b.Start()
}
