# Cubic Spacy Bot

[![Build](https://github.com/database64128/cubic-spacy-bot/actions/workflows/build.yml/badge.svg)](https://github.com/database64128/cubic-spacy-bot/actions/workflows/build.yml)
[![Release](https://github.com/database64128/cubic-spacy-bot/actions/workflows/release.yml/badge.svg)](https://github.com/database64128/cubic-spacy-bot/actions/workflows/release.yml)
[![AUR version](https://img.shields.io/aur/version/cubic-spacy-bot-git?label=cubic-spacy-bot-git)](https://aur.archlinux.org/packages/cubic-spacy-bot-git)

An inline Telegram bot that gives you plenty of space!

## Deployment

```console
$ git clone https://github.com/database64128/cubic-spacy-bot.git
$ cd cubic-spacy-bot/
$ go build -trimpath -ldflags '-s -w'
$ sudo ln -rs cubic-spacy-bot /usr/bin/
$ sudo ln -rs systemd/system/cubic-spacy-bot.service /usr/lib/systemd/system/
$ sudo mkdir /etc/cubic-spacy-bot
$ sudo nano /etc/cubic-spacy-bot/env
$ sudo systemctl enable --now cubic-spacy-bot.service
```

Add the following when editing `/etc/cubic-spacy-bot/env`:

```bash
TELEGRAM_BOT_TOKEN=1234567:4TT8bAc8GHUspu3ERYn-KGcvsvGB9u_n4ddy
```

To use webhooks, specify additional environment variables:

```bash
TELEGRAM_BOT_WEBHOOK_LISTEN_NETWORK=unix
TELEGRAM_BOT_WEBHOOK_LISTEN_ADDRESS=/run/cubic-spacy-bot.sock
# TELEGRAM_BOT_WEBHOOK_LISTEN_MODE=0777
# TELEGRAM_BOT_WEBHOOK_SECRET_TOKEN=secret_token
TELEGRAM_BOT_WEBHOOK_URL=https://example.com/cubic-spacy-bot
```

## License

[AGPLv3](LICENSE)
