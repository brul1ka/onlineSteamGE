package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleAppRequest(name string) (string, []string, []*AppItem) {
	matches, suggestions := searchAppByName(name)

	if len(matches) > 0 {
		if len(matches) == 1 {
			return getAppOnline(matches[0]), nil, matches
		}
		return "Found several games with this name. Please choose:", nil, matches
	}

	if len(suggestions) > 0 {
		return "Game not found. Maybe you meant:", suggestions, nil
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
	chatID := cb.Message.Chat.ID

	bot.Request(tgbotapi.NewCallback(cb.ID, ""))

	if strings.HasPrefix(data, "refresh_") || strings.HasPrefix(data, "id_") {
		var idStr string
		if strings.HasPrefix(data, "refresh_") {
			idStr = strings.TrimPrefix(data, "refresh_")
		} else {
			idStr = strings.TrimPrefix(data, "id_")
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			log.Printf("Error converting ID: %v", err)
			return
		}

		app := searchAppByID(id)
		if app != nil {
			msgText := getAppOnline(app)
			sendGameWithPhoto(bot, chatID, app, msgText)
		} else {
			msg := tgbotapi.NewMessage(chatID, "⚠️ Game info lost. Please search again.")
			bot.Send(msg)
		}
		return
	}

	msgText, suggestions, foundApps := handleAppRequest(data)

	if len(foundApps) > 0 {
		if len(foundApps) == 1 {
			sendGameWithPhoto(bot, chatID, foundApps[0], msgText)
		} else {
			reply := tgbotapi.NewMessage(chatID, msgText)
			reply.ReplyMarkup = createMatchesKeyboard(foundApps)
			bot.Send(reply)
		}
		return
	}

	if len(suggestions) > 0 {
		msg := tgbotapi.NewMessage(chatID, msgText)
		msg.ReplyMarkup = createSuggestionsKeyboard(suggestions)
		bot.Send(msg)
		return
	}

	if msgText != "" {
		msg := tgbotapi.NewMessage(chatID, msgText)
		msg.ParseMode = "HTML"
		bot.Send(msg)
	}
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
	log.Printf("[%s] %s", msg.From.UserName, msg.Text)

	botMsgText, suggestions, foundApps := handleAppRequest(msg.Text)

	if len(foundApps) > 0 {
		if len(foundApps) == 1 {
			sendGameWithPhoto(bot, chatID, foundApps[0], botMsgText)
		} else {
			reply := tgbotapi.NewMessage(chatID, botMsgText)
			reply.ReplyMarkup = createMatchesKeyboard(foundApps)
			bot.Send(reply)
		}
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
