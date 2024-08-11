# Writeup Finder

Writeup Finder is a tool designed to automatically find and save recent writeups from specified URLs. It supports saving the found writeups in a JSON file, a PostgreSQL database, or sending them directly to a Telegram channel.

## Features

- Fetch recent writeups from multiple URLs.
- Save writeups to a JSON file or a PostgreSQL database.
- Optionally send notifications of new writeups to a Telegram channel.
- Configurable through flags for different output methods.

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
    ```
4. Update the `url.txt` file with the URLs you want to monitor.
5. Run the tool with the desired flags.

## Usage

```bash
go run main.go -f  # Save to JSON file
go run main.go -d  # Save to PostgreSQL database
go run main.go -t  # Send new writeups to Telegram
```


You can use Cron for to run script every *hours, *days, or etc.
