# Steam Online: Golang Edition - A Telegram Bot for Checking Game Player Counts
**Steam Online: GE** is Telegram bot for checking games and it's **player count** right now. This bot is fast, simple and minimalistic.
## Features
* **All games** - you can check any game, and it will be here.
* **Thumbnails** - if available, an image of the game will be attached to the message.
* **More info** - if you want learn more about a particular game, the message includes a link to SteamDB for this game!
## Usage
Just type in the game's name.

If you want to search for a game using a specific keyword (even if it matches an existing game's name), use the `/find <keyword>` command.

## Installation & Run

1. Clone the repository:
   ```bash
   git clone https://github.com/brul1ka/onlineSteamGE.git
   cd onlineSteamGE
   ```
2. Create a `.env` file in the root directory and add your tokens just as shown in `.env.example`
3. Download dependencies:
    ```bash
    go mod download
    ```
4. Build and run the bot:
    ```bash
    go run main.go
    ```
### Running with Docker

If you prefer not to install Go locally, you can easily run the bot using Docker and Docker Compose.

1. Make sure you have **Docker** and **Docker Compose** installed.
2. Fill in your credentials in the `.env` file.
3. Start the bot in detached mode:

```bash
docker compose up -d