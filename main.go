package main

import (
	"log"
	"os"
	"strconv"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

var (
	apps  []App
	mutex sync.RWMutex
)

func main() {
	godotenv.Load()
	token := os.Getenv("ONLINESTEAMGE_BOT_TOKEN")
	adminChatIDstr := os.Getenv("ADMIN_CHAT_ID")
	if token == "" {
		log.Panic("Didn't find bot token in environment variable!")
	}
	if adminChatIDstr == "" {
		log.Panic("Didn't find bot admin chat ID in environment variable!")
	}
	adminChatID, err := strconv.Atoi(adminChatIDstr)
	if err != nil {
		log.Panicf("Can't convert admin chat id from str to int: %v", err)
	}
	initDB()

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Authorized as: %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	getAppList()
	go setupCron()

	publicScope := tgbotapi.NewBotCommandScopeDefault()
	setPublicCommands := tgbotapi.NewSetMyCommandsWithScope(publicScope, publicCommands...)
	_, err = bot.Request(setPublicCommands)
	if err != nil {
		log.Fatalf("Error setting public command menu: %v", err)
	}

	adminScope := tgbotapi.NewBotCommandScopeChat(int64(adminChatID))
	setAdminCommands := tgbotapi.NewSetMyCommandsWithScope(adminScope, adminCommands...)
	_, err = bot.Request(setAdminCommands)
	if err != nil {
		log.Fatalf("Error setting admin command menu: %v", err)
	} else {
		log.Println("Command menus setted successfully")
	}

	for update := range updates {
		if update.CallbackQuery != nil {
			log.Printf("[\"%s\" requested callback] %s", update.CallbackQuery.From.UserName, update.CallbackQuery.Data)
			go handleCallback(bot, update.CallbackQuery)
			continue
		}
		if update.Message != nil {
			go handleMessage(bot, update.Message)
		}
	}
}
