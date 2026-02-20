package main

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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
				welcomeMsg := tgbotapi.NewMessage(chatID, "Hello! It's Online Steam, TG-bot to check online in Steam game!\nEnter appid to check it's online.")
				bot.Send(welcomeMsg)
				continue
			}
		}
		log.Printf("[%s] %s", userName, msgTxt)
		//
		//
		//
		userAppName := msgTxt
		app, err := getAppByName(userAppName)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("error! %v", err)))
		}
		bot.Send(tgbotapi.NewMessage(chatID, GetSteamOnline(app)))
	}
}
