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
		log.Fatal("[FATAL] did NOT find bot token in environment variable!")
	}
	if adminChatIDstr == "" {
		log.Fatal("[FATAL] did NOT find bot admin chat ID in environment variable!")
	}
	adminChatID, err := strconv.Atoi(adminChatIDstr)
	if err != nil {
		log.Fatalf("[FATAL] cant convert admin chat id from str to int: %v", err)
	}
	initDB()

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("[FATAL] failed to initialize bot: %v", err)
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
		log.Printf("[ERROR] failed to set public command menu: %v", err)
	}

	adminScope := tgbotapi.NewBotCommandScopeChat(int64(adminChatID))
	setAdminCommands := tgbotapi.NewSetMyCommandsWithScope(adminScope, adminCommands...)
	_, err = bot.Request(setAdminCommands)
	if err != nil {
		log.Printf("[ERROR] failed to set admin command menu: %v", err)
	} else {
		log.Println("[SUCCESS] command menus setted successfully")
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
