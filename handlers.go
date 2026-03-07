package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleAppRequest(name string) (string, []string, *AppItem) {
	mutex.RLock()
	defer mutex.RUnlock()
	app, suggestions := searchAppByName(name)

	if app != nil {
		return getAppOnline(app), nil, app
	}
	if len(suggestions) > 0 {
		if len(suggestions) == 1 {
			log.Printf("Auto-finding game: %s", suggestions[0])
			autoGame, _ := searchAppByName(suggestions[0])
			return fmt.Sprintf("🪄%s", getAppOnline(autoGame)), nil, autoGame
		}
		return "No game with this name! Maybe you meant:", suggestions, nil
	}
	return "There is no such game with this name!", nil, nil
}

func handleSearchRequest(name string) (string, []string) {
	mutex.RLock()
	defer mutex.RUnlock()

	suggestions := filterAppsByQuery(name, &apps)

	if len(suggestions) > 0 {
		return fmt.Sprintf("🔍 Search results for '%s':", name), suggestions
	}
	return "Nothing found for your request!", nil
}

func handleCallback(bot *tgbotapi.BotAPI, cb *tgbotapi.CallbackQuery) {
	data := cb.Data
	var msgText string
	var app *AppItem

	bot.Request(tgbotapi.NewCallback(cb.ID, ""))

	if strings.HasPrefix(data, "refresh_") {
		idStr := strings.TrimPrefix(data, "refresh_")
		id, _ := strconv.Atoi(idStr)
		app = findAppByID(id)
		if app != nil {
			msgText = getAppOnline(app)
		} else {
			msgText = "Game info lost. Please search again."
		}
	} else {
		msgText, _, app = handleAppRequest(data)
	}

	if app != nil {
		sendGameWithPhoto(bot, cb.Message.Chat.ID, app, msgText)
		return
	}
	msg := tgbotapi.NewMessage(cb.Message.Chat.ID, msgText)
	msg.ParseMode = "HTML"
	bot.Send(msg)
}

func handleCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID

	switch msg.Command() {
	case "start":
		text := "Hello! It's Online Steam, TG-bot to check online in Steam game!\nJust type name of the game you wanna check."
		bot.Send(tgbotapi.NewMessage(chatID, text))

	case "find":
		query := msg.CommandArguments()
		if query == "" {
			bot.Send(tgbotapi.NewMessage(chatID, "Please enter game name after /find"))
			return
		}
		msgText, suggestions := handleSearchRequest(query)
		reply := tgbotapi.NewMessage(chatID, msgText)
		if len(suggestions) > 0 {
			reply.ReplyMarkup = createSuggestionsKeyboard(suggestions)
		}
		bot.Send(reply)
	}
}

func handleMessage(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	if msg.IsCommand() {
		handleCommand(bot, msg)
		return
	}

	chatID := msg.Chat.ID
	if msg.Text == "" {
		bot.Send(tgbotapi.NewMessage(chatID, "Not valid message!"))
		return
	}

	log.Printf("[%s] %s", msg.From.UserName, msg.Text)
	botMsgText, suggestions, foundApp := handleAppRequest(msg.Text)

	if foundApp != nil {
		sendGameWithPhoto(bot, chatID, foundApp, botMsgText)
		return
	}

	if len(suggestions) > 0 {
		botMsg := tgbotapi.NewMessage(chatID, botMsgText)
		botMsg.ReplyMarkup = createSuggestionsKeyboard(suggestions)
		bot.Send(botMsg)
		return
	}

	bot.Send(tgbotapi.NewMessage(chatID, botMsgText))
}
