package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type SteamResponse struct {
	Response struct {
		PlayerCount int `json:"player_count"`
		Result      int `json:"result"`
	} `json:"response"`
}

func main() {

	bot, err := tgbotapi.NewBotAPI(TOKEN)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Authorized as: %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		msgTxt := update.Message.Text
		userName := update.Message.From.UserName
		chatID := update.Message.Chat.ID

		if update.Message.IsCommand() {
			if update.Message.Command() == "start" {
				log.Printf("User %s started bot: %s", userName, msgTxt)
				msg := tgbotapi.NewMessage(chatID, "Hello! It's Online Steam, TG-bot to check online in Steam game!\nEnter appid to check it's online.")
				bot.Send(msg)
				continue
			}
		}
		log.Printf("[%s] %s", userName, msgTxt)
		//
		//
		//
		appID, err := strconv.Atoi(msgTxt)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "Enter correct integer please"))
			continue
		}

		url := fmt.Sprintf("https://api.steampowered.com/ISteamUserStats/GetNumberOfCurrentPlayers/v1/?appid=%d", appID)
		resp, err := http.Get(url)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "Error with connecting to Steam API!"))
			continue
		}
		defer resp.Body.Close()

		var data SteamResponse
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Error trying to parse"))
			continue
		}
		if data.Response.Result != 1 {
			bot.Send(tgbotapi.NewMessage(chatID, "Game with this appID not found!"))
			continue
		}
		finalMsg := fmt.Sprintf("📊 Now in game (AppID %d): %d people", appID, data.Response.PlayerCount)
		bot.Send(tgbotapi.NewMessage(chatID, finalMsg))

	}
}
