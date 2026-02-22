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
			if update.Message.Command() == "start" {
				log.Printf("User %s started bot: %s", userName, msgTxt)
				welcomeMsg := tgbotapi.NewMessage(chatID, "Hello! It's Online Steam, TG-bot to check online in Steam game!\nEnter appid to check it's online.")
				bot.Send(welcomeMsg)
				continue
			}
		}
		log.Printf("[%s] %s", userName, msgTxt)

		botMsgText, suggestions := handleAppRequest(msgTxt)
		botMsg := tgbotapi.NewMessage(chatID, botMsgText)
		if len(suggestions) > 0 {
			var rows [][]tgbotapi.InlineKeyboardButton
			for _, name := range suggestions {
				callbackData := name
				if len(callbackData) > 64 {
					callbackData = callbackData[:64]
				}
				btn := tgbotapi.NewInlineKeyboardButtonData(name, callbackData)
				rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
			}
			botMsg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
		}
		bot.Send(botMsg)
		log.Printf("[%s] %s", bot.Self.UserName, botMsgText)
	}
}
