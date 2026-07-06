package main

import (
	"log"
	"os"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

var (
	apps  Apps
	mutex sync.RWMutex
)

func main() {
	godotenv.Load()
	token := os.Getenv("ONLINESTEAMGE_BOT_TOKEN")
	if token == "" {
		log.Panic("Didn't find bot token in environment variable!")
	}
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
