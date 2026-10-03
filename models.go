package main

import (
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type SteamResponse struct {
	Response struct {
		PlayerCount int `json:"player_count"`
		Result      int `json:"result"`
	} `json:"response"`
}

type App struct {
	AppID             int    `json:"appid"`
	Name              string `json:"name"`
	LastModified      int    `json:"last_modified"`
	PriceChangeNumber int    `json:"price_change_number"`
}

type AppDetails struct {
	Success bool `json:"success"`
	Data    struct {
		ReleaseDate struct {
			ComingSoon bool   `json:"coming_soon"`
			Date       string `json:"date"`
		} `json:"release_date"`
	} `json:"data"`
}

type GameStat struct {
	AppID       int
	OnlineCount int
	CheckedAt   time.Time
}

// commands available for all users
var publicCommands []tgbotapi.BotCommand = []tgbotapi.BotCommand{
	{
		Command:     "start",
		Description: "start bot and get greetings",
	},
	{
		Command:     "find",
		Description: "force find games containing this substring",
	},
	{
		Command:     "set_timezone",
		Description: "Set timezone (format: +HH:MM or -HH:MM)",
	},
}

// commands for admins
var adminCommands = append(
	// append public commands to empty list and then to this new public commands list append admin commands
	append([]tgbotapi.BotCommand{}, publicCommands...),
	tgbotapi.BotCommand{
		Command:     "broadcast",
		Description: "ADMIN: send message to all bot users",
	},
)
