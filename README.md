# Writeup Finder

[![Go Version](https://img.shields.io/badge/go-1.17%20%7C%201.18%20%7C%201.19%20%7C%201.20-blue)](https://golang.org/dl/)
[![GitHub Issues](https://img.shields.io/github/issues/mamad-1999/writeup-finder)](https://github.com/mamad-1999/writeup-finder/issues)
[![GitHub Stars](https://img.shields.io/github/stars/mamad-1999/writeup-finder)](https://github.com/mamad-1999/writeup-finder/stargazers)
[![GitHub License](https://img.shields.io/github/license/mamad-1999/writeup-finder)](https://github.com/mamad-1999/writeup-finder/blob/master/LICENSE)

<p>
    <a href="https://skillicons.dev">
      <img src="https://github.com/tandpfun/skill-icons/blob/main/icons/GoLang.svg" width="48" title="Go">
      <img src="https://github.com/tandpfun/skill-icons/blob/main/icons/Github-Dark.svg" width="48" title="github">
    </a>
</p>

Writeup Finder is a tool designed to automatically find and save recent writeups from specified URLs. It supports saving the found writeups in a JSON file, a PostgreSQL database, or sending them directly to a Telegram channel.

## Features

- Fetch recent writeups from multiple URLs.
- Save writeups to a JSON file or a PostgreSQL database.
- Optionally send notifications of new writeups to a Telegram channel.
- Configurable through flags for different output methods.


```
├── .env
├── .gitignore
├── README.md
├── config/
│   ├── env.go
├── data/
│   ├── found-url.json
│   ├── url.txt
├── db/
│   ├── db.go
├── directory_structure.md
├── go.mod
├── go.sum
├── main.go
├── rss/
│   ├── fetch.go
├── telegram/
│   ├── telegram.go
├── utils/
    ├── http.go
    ├── utils.go
```

## Requirements

- Go 1.16+
- PostgreSQL
- A Telegram bot (for sending notifications)

## Setup

1. Clone the repository.
2. Install dependencies using `go mod tidy`.
3. Create a `.env` file with the following variables:
    ```bash
    DB_USER=<your_db_user>
    DB_PASSWORD=<your_db_password>
    DB_HOST=<your_db_host>
    DB_PORT=<your_db_port>
    DB_NAME=<your_db_name>
    
    TELEGRAM_BOT_TOKEN=<your_telegram_bot_token>
    TELEGRAM_CHANNEL_ID=<your_telegram_channel_id>
    CHAT_ID=<CHAT_ID> # in the group
    MESSAGE_THREAD_ID=<MESSAGE_TREAD_ID> # superGroup Topic
    ```
4. Update the `url.txt` file with the URLs you want to monitor.
5. Run the tool with the desired flags.

## Usage

```bash
go run main.go -f  # Save to JSON file
go run main.go -d  # Save to PostgreSQL database
go run main.go [-d / -f] -t  # Send new writeups to Telegram
go run main.go [-d / -f] -t --proxy=PROTOCOL://HOSTNAME:PORT #proxy just work for telegram send message [-t]
```

You can use `CRON` to run script every *hours, *days, or etc.
