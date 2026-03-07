package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/robfig/cron/v3"
)

// FUNCTIONS FOR OPERATIONS WITH APP

func searchAppByName(name string) (*AppItem, []string) {
	mutex.RLock()
	defer mutex.RUnlock()
	cleanName := strings.TrimSpace(name)

	for i := range apps {
		if strings.EqualFold(strings.TrimSpace(apps[i].Name), cleanName) {
			return &apps[i], nil
		}
	}
	filtered := filterAppsByQuery(name, &apps)
	return nil, filtered
}

func getAppOnline(app *AppItem) string {
	if app == nil {
		return "There is no such game with this name!"
	}
	if app.Name == "" {
		return "Not valid message"
	}

	url := fmt.Sprintf("https://api.steampowered.com/ISteamUserStats/GetNumberOfCurrentPlayers/v1/?appid=%d", app.AppID)
	resp, err := http.Get(url)
	if err != nil {
		return "Error connecting to Steam API!"
	}
	defer resp.Body.Close()

	var data SteamResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "Error parsing response!"
	}

	if data.Response.Result != 1 {
		return "Failed to retrieve online!\nMaybe this game isn't released"
	}

	return fmt.Sprintf("📊 Now in game %s (appid: %d): %d people", app.Name, app.AppID, data.Response.PlayerCount)
}

func filterAppsByQuery(query string, apps *Apps) []string {
	result := []string{}
	queryWords := strings.Fields(strings.ToLower(query))

	for _, app := range *apps {
		counter := 0
		appName := strings.ToLower(app.Name)

		for _, word := range queryWords {
			if strings.Contains(appName, word) {
				counter++
			}
		}

		if counter == len(queryWords) {
			result = append(result, app.Name)
		}
		if len(result) >= 20 {
			break
		}
	}
	return result
}

func findAppByID(id int) *AppItem {
	mutex.RLock()
	defer mutex.RUnlock()
	for i := range apps {
		if apps[i].AppID == id {
			return &apps[i]
		}
	}
	return nil
}

func sendGameWithPhoto(bot *tgbotapi.BotAPI, chatID int64, app *AppItem, text string) {
	photoURL := fmt.Sprintf("https://cdn.akamai.steamstatic.com/steam/apps/%d/header.jpg", app.AppID)
	photoMsg := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(photoURL))

	photoMsg.Caption = text
	keyboard := createCheckAgainKeyboard(app.AppID)
	photoMsg.ReplyMarkup = keyboard

	bot.Send(photoMsg)
}

// FUNCTIONS FOR WORKING WITH INLINE KEYBOARD

func createSuggestionsKeyboard(suggestions []string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, name := range suggestions {
		data := name
		if len(data) > 64 {
			data = data[:64]
		}
		btn := tgbotapi.NewInlineKeyboardButtonData(name, data)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func createCheckAgainKeyboard(appID int) tgbotapi.InlineKeyboardMarkup {
	callbackData := fmt.Sprintf("refresh_%d", appID)

	btn := tgbotapi.NewInlineKeyboardButtonData("🔄 Check Again", callbackData)
	return tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(btn))
}

// APP LIST FUNCTIONS

func loadApps() *Apps {
	file, err := os.Open("games_appid.json")
	if err != nil {
		return nil
	}
	defer file.Close()

	var apps Apps
	if err := json.NewDecoder(file).Decode(&apps); err != nil {
		return &apps
	}
	return &apps
}

func getAppList() {
	url := "https://raw.githubusercontent.com/jsnli/steamappidlist/refs/heads/master/data/games_appid.json"
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Failed to get app list: %v", err)
		return
	}
	defer resp.Body.Close()

	file, err := os.Create("temp.json")
	if err != nil {
		log.Printf("Failed to create game list file: %v", err)
		return
	}

	bytesWritten, err := io.Copy(file, resp.Body)
	if err != nil {
		file.Close()
		log.Printf("Failed to copy stream to file: %v", err)
		return
	}
	file.Close()

	err = os.Rename("temp.json", "games_appid.json")
	if err != nil {
		log.Printf("Failed to rename file: %v", err)
		return
	}
	os.Rename("temp.json", "games_appid.json")

	loaded := loadApps()
	if loaded != nil {
		mutex.Lock()
		apps = *loaded
		mutex.Unlock()
	}

	log.Printf("Successfully got an app list. Bytes written: %d", bytesWritten)
}

func updateAppList() {
	c := cron.New()

	_, err := c.AddFunc("0 3 * * *", getAppList)
	if err != nil {
		log.Printf("Failed to configure cron: %v", err)
		return
	}

	c.Start()
	log.Print("cron started")
}
