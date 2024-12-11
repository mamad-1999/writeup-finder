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

"Join our Writeup Hacking supergroup for curated hacking writeups and resources! 📜🔍 https://t.me/writeup_hacking"

Writeup Finder is a tool designed to automatically find and save recent writeups from specified URLs. It supports saving the found writeups in a PostgreSQL database, and sending them directly to a Telegram channel Or group.

## Features

- Fetch recent writeups from multiple URLs.
- Save writeups to a PostgreSQL database.
- Optionally send notifications of new writeups to a Telegram channel.
- Configurable through flags for different output methods.

```
# Directory structure of WriteUp-finder

├── .env 
├── .env.example 
├── .gitignore 
├── README.md 
├── command/ 
│   └── command.go 
├── config/ 
│   └── env.go 
├── data/ 
│   ├── keywords.json 
│   └── url.txt 
├── db/ 
│   └── db.go 
├── global/ 
│   └── global.go 
├── go.mod 
├── go.sum 
├── main.go 
├── rss/ 
│   └── fetch.go 
├── run_writeUp-finder.sh 
├── telegram/ 
│   └── telegram.go 
├── url/ 
│   └── url.go 
├── utils/ 
│   ├── filters.go 
│   ├── http.go 
│   └── utils.go 
└── writeup-finder 
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
   # Database Configuration
   DB_USER=<your_db_user>             # Your database username
   DB_PASSWORD=<your_db_password>     # Your database password
   DB_HOST=<your_db_host>             # The hostname of your database server
   DB_PORT=<your_db_port>             # The port number for your database connection
   DB_NAME=<your_db_name>             # The name of your database

   # Telegram Configuration
   TELEGRAM_BOT_TOKEN=<your_telegram_bot_token>      # Your Telegram bot's token
   TELEGRAM_CHANNEL_ID=<your_telegram_channel_id>    # The ID of your Telegram channel
   CHAT_ID=<CHAT_ID>                                 # The ID of the chat group
   MESSAGE_THREAD_ID=<MESSAGE_THREAD_ID>             # The ID of the message thread (for supergroups with topics)
   ```

4. Update the `url.txt` file with the URLs you want to monitor.
5. Run the tool with the desired flags.

## Usage

| Command                                                   | Description                         |
| --------------------------------------------------------- | ----------------------------------- |
| `go run main.go -d`                                       | Save to PostgreSQL database         |
| `go run main.go [-d] -t`                                  | Send new writeups to Telegram       |
| `go run main.go [-d] -t --proxy=PROTOCOL://HOSTNAME:PORT` | Send to Telegram with proxy support |

You can use `CRON` to run script every *hours, *days, or etc.

#### Example for run script every 3 hour

More read: [How to Automate Tasks with cron Jobs in Linux](https://www.freecodecamp.org/news/cron-jobs-in-linux/)

```bash
    0 */3 * * * cd /path/to/your/script && /usr/local/go/bin/go run /path/to/your/project/main.go -d -t
```
