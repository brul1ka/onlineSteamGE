package main

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleAppRequest(name string) (string, []string) {
	mutex.RLock()
	defer mutex.RUnlock()
	app, suggestions := searchAppByName(name)

	if app != nil {
		return getAppOnline(app), nil
	}
	if len(suggestions) > 0 {
		if len(suggestions) == 1 {
			autoGame, _ := searchAppByName(suggestions[0])
			return fmt.Sprintf("🪄 Auto found!\n%s", getAppOnline(autoGame)), nil
		}
		return "No game with this name! Maybe you meant:", suggestions
	}
	return "There is no such game with this name!", nil
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
	gameName := cb.Data
	msgText, _ := handleAppRequest(gameName)

	bot.Request(tgbotapi.NewCallback(cb.ID, ""))
	msg := tgbotapi.NewMessage(cb.Message.Chat.ID, msgText)
	bot.Send(msg)
}

func handleCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID

	switch msg.Command() {
	case "start":
		text := `Hello! It's Online Steam, TG-bot to check online in Steam game!
		Just type name of the game you wanna check.`
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
			reply.ReplyMarkup = createSuggestInlineKeyboard(suggestions)
		}
		bot.Send(reply)
	}
}

func handleMessage(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	if msg.IsCommand() {
		handleCommand(bot, msg)
		return
	}

	msgTxt := msg.Text
	chatID := msg.Chat.ID
	userName := msg.From.UserName

	log.Printf("[%s] %s", userName, msgTxt)
	botMsgText, suggestions := handleAppRequest(msgTxt)
	botMsg := tgbotapi.NewMessage(chatID, botMsgText)
	if len(suggestions) > 0 {
		botMsg.ReplyMarkup = createSuggestInlineKeyboard(suggestions)
	}

	bot.Send(botMsg)
}
