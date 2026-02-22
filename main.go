package main

import (
	"fmt"
	"log"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	apps  Apps
	mutex sync.RWMutex
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

	getAppList()
	go updateAppList()

	for update := range updates {
		if update.CallbackQuery != nil {
			gameName := update.CallbackQuery.Data
			msgText, _ := handleAppRequest(gameName)

			callbackMsg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, msgText)
			bot.Send(callbackMsg)
			callbackConfig := tgbotapi.NewCallback(update.CallbackQuery.ID, fmt.Sprintf("FOUND: %s", gameName))
			bot.Request(callbackConfig)

			continue
		}
		if update.Message == nil {
			continue
		}

		msgTxt := update.Message.Text
		userName := update.Message.From.UserName
		chatID := update.Message.Chat.ID

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				log.Printf("User %s started bot: %s", userName, msgTxt)
				welcomeMsg := tgbotapi.NewMessage(chatID, "Hello! It's Online Steam, TG-bot to check online in Steam game!\nJust type name of the game you wanna check.")
				bot.Send(welcomeMsg)
				continue
			case "find":
				query := update.Message.CommandArguments()
				if query == "" {
					bot.Send(tgbotapi.NewMessage(chatID, "Please enter game name after /find"))
					continue
				}
				msgText, suggestions := handleSearchRequest(query)
				msg := tgbotapi.NewMessage(chatID, msgText)
				if len(suggestions) > 0 {
					msg.ReplyMarkup = createInlineKeyboard(suggestions)
				}
				bot.Send(msg)
			}
			continue
		}
		log.Printf("[%s] %s", userName, msgTxt)

		botMsgText, suggestions := handleAppRequest(msgTxt)
		botMsg := tgbotapi.NewMessage(chatID, botMsgText)

		if len(suggestions) > 0 {
			botMsg.ReplyMarkup = createInlineKeyboard(suggestions)
		}

		bot.Send(botMsg)
	}
}
