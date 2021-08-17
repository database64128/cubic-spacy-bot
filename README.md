# Cubic Spacy Bot

An inline Telegram bot that gives you plenty of space!

## Deployment

```console
$ git clone https://github.com/database64128/cubic-spacy-bot.git
$ cd cubic-spacy-bot/
$ go build -trimpath -ldflags '-s -w -buildid='
$ sudo ln -rs cubic-spacy-bot /usr/bin/
$ sudo ln -rs systemd/cubic-spacy-bot.service /usr/lib/systemd/system/
$ sudo systemctl edit cubic-spacy-bot.service
$ sudo systemctl enable --now cubic-spacy-bot.service
```

Add the following when editing the service:

```systemd
[Service]
Environment=TELEGRAM_BOT_TOKEN=1234567:4TT8bAc8GHUspu3ERYn-KGcvsvGB9u_n4ddy
```

The service unit file can also be used as a user unit.

## License

[AGPLv3](LICENSE)
