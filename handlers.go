package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleAppRequest(name string) (string, []string, []*App) {
	matches, suggestions := getAppsByName(name)

	if len(matches) > 0 {
		if len(matches) == 1 {
			return formatAppOnlineMessage(matches[0], false), nil, matches
		}
		return "Found several games with this name. Please choose:", nil, matches
	}

	if len(suggestions) > 0 {
		if len(suggestions) == 1 {
			app, _ := getAppsByName(suggestions[0])

			// check that getAppsByName actually found something
			// and returned exactly one game (in case there are no duplicates under that name)
			if len(app) == 1 {
				return formatAppOnlineMessage(app[0], true), nil, app
			}

			// If there are multiple apps hidden under this "single" name in the database,
			// we can't show one card—we present them as exact matches,
			// and the bot will build a MatchesKeyboard for them!
			if len(app) > 1 {
				return "Found several games with this name. Please choose:", nil, app
			}
		}
		return "Game not found. Maybe you meant:", suggestions, nil
	}
	return "There is no such game with this name!", nil, nil
}

func handleForceFindRequest(name string) (string, []string) {
	mutex.RLock()
	defer mutex.RUnlock()

	suggestions := fuzzySearchApps(name, &apps)

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

		app := getAppByID(id)
		if app != nil {
			msgText := formatAppOnlineMessage(app, false)
			sendFinalMessage(bot, chatID, app, msgText)
		} else {
			msg := tgbotapi.NewMessage(chatID, "⚠️ Game info lost. Please search again.")
			_, err := bot.Send(msg)
			if err != nil {
				log.Printf("Error sending game info lost message: %v", err)
			}
		}
		return
	}

	msgText, suggestions, foundApps := handleAppRequest(data)

	if len(foundApps) > 0 {
		if len(foundApps) == 1 {
			sendFinalMessage(bot, chatID, foundApps[0], msgText)
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
	// all users
	case "start":
		text := "Hello! It's Online Steam, TG-bot to check online in Steam game!\nJust type name of the game you wanna check."
		bot.Send(tgbotapi.NewMessage(chatID, text))

	case "find":
		query := msg.CommandArguments()
		if query == "" {
			bot.Send(tgbotapi.NewMessage(chatID, "Please enter game name after /find"))
			return
		}
		msgText, suggestions := handleForceFindRequest(query)
		reply := tgbotapi.NewMessage(chatID, msgText)
		if len(suggestions) > 0 {
			reply.ReplyMarkup = createSuggestionsKeyboard(suggestions)
		}
		bot.Send(reply)

	// admin
	case "broadcast":
		text := msg.CommandArguments()

		query := `SELECT chat_id FROM chats`
		rows, err := db.Query(query)
		if err != nil {
			log.Printf("error getting chats from db: %v\n", err)
		}
		defer rows.Close()

		var chats []int64

		for rows.Next() {
			var c int64
			err := rows.Scan(
				&c,
			)
			if err != nil {
				log.Println("Error reading line:", err)
				continue
			}

			chats = append(chats, c)
		}
		if err := rows.Err(); err != nil {
			log.Printf("Error reading db: %v", err)
		}

		for _, chatID := range chats {
			message := tgbotapi.NewMessage(chatID, text)
			bot.Send(message)
		}
	}
}

func handleMessage(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	log.Printf("[%s] %s", msg.From.UserName, msg.Text)

	query := `INSERT OR IGNORE INTO chats (chat_id) VALUES (?);`
	_, err := db.Exec(query, chatID)
	if err != nil {
		log.Printf("Error saving chat %d into database: %v", chatID, err)
	}

	if msg.IsCommand() {
		handleCommand(bot, msg)
		return
	}

	botMsgText, suggestions, foundApps := handleAppRequest(msg.Text)

	if len(foundApps) > 0 {
		if len(foundApps) == 1 {
			sendFinalMessage(bot, chatID, foundApps[0], botMsgText)
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
