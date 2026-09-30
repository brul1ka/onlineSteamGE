package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/robfig/cron/v3"
)

// FUNCTIONS FOR OPERATIONS WITH APP

func getAppsByName(name string) ([]*App, []string) {
	mutex.RLock()
	defer mutex.RUnlock()
	cleanName := strings.TrimSpace(name)

	var exactMatches []*App
	for i := range apps {
		if strings.EqualFold(strings.TrimSpace(apps[i].Name), cleanName) {
			exactMatches = append(exactMatches, &apps[i])
		}
	}

	if len(exactMatches) > 0 {
		return exactMatches, nil
	}

	filtered := fuzzySearchApps(name, &apps)
	return nil, filtered
}

func formatAppOnlineMessage(app *App, isSingleMatch bool) string {
	if app == nil {
		return "There is no such game with this name!"
	}
	if app.Name == "" || app.Name == " " {
		return "Not valid message"
	}

	// get number of online rn in app
	url := fmt.Sprintf("https://api.steampowered.com/ISteamUserStats/GetNumberOfCurrentPlayers/v1/?appid=%d", app.AppID)
	resp, err := http.Get(url)
	if err != nil {
		return "<b>Error connecting to Steam API!</b>"
	}
	defer resp.Body.Close()

	var data SteamResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "Error parsing response!"
	}

	if data.Response.Result != 1 {
		return fmt.Sprintf("<b>Failed to retrieve online!</b>\nMaybe this game isn't released\n\nSteamDB page about this game is <a href='https://steamdb.info/app/%d/charts/'>here</a>", app.AppID)
	}

	// end of getting online, now we getting release date
	url = fmt.Sprintf("https://store.steampowered.com/api/appdetails?appids=%d&filters=release_date", app.AppID)
	resp, err = http.Get(url)
	if err != nil {
		return "<b>Error connecting to Steam API!</b>"
	}
	defer resp.Body.Close()

	appIDStr := strconv.Itoa(app.AppID)
	var result map[string]AppDetails
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "Error parsing response!"
	}

	data2, exists := result[appIDStr]
	releaseDate := "Release date is unknown"

	if exists && data2.Success {
		if data2.Data.ReleaseDate.ComingSoon {
			releaseDate = "Coming soon..."
		} else if data2.Data.ReleaseDate.Date != "" {
			releaseDate = data2.Data.ReleaseDate.Date
		}
	}

	online_count := data.Response.PlayerCount
	query1 := `INSERT INTO game_stats (app_id, online_count) VALUES (?, ?);`
	_, err = db.Exec(query1, app.AppID, online_count)
	if err != nil {
		log.Printf("[ERROR] failed to insert game online stats into database: %v", err)
	}

	query2 := `SELECT app_id, online_count, checked_at 
			FROM game_stats 
			WHERE app_id = ?
			AND checked_at >= datetime('now', '-1 day', '-1 hour')
			AND checked_at <= datetime('now', '-1 day', '+1 hour')
			ORDER BY checked_at DESC;
			`
	var gamestat GameStat
	err = db.QueryRow(query2, app.AppID).Scan(&gamestat.AppID, &gamestat.OnlineCount, &gamestat.CheckedAt)

	if err != nil {
		if err == sql.ErrNoRows {

		} else {
			log.Printf("[ERROR] failed to get app online yesterday: %v", err)
		}
	}
	successTxt := fmt.Sprintf("\n📊 Now in game <code>%s</code> (%s; ID: <a href='https://steamdb.info/app/%d/charts/'>%d</a>): %d people\n", app.Name, releaseDate, app.AppID, app.AppID, online_count)
	if isSingleMatch {
		successTxt = "🪄" + successTxt
	}
	if gamestat.AppID != 0 {
		t := gamestat.CheckedAt.Format("06/01/02 15:04:05")
		yesterday := fmt.Sprintf("❗️Yesterday, at approximately this time (UTC+00:00 %s), the online for this game was %d people.", t, gamestat.OnlineCount)
		successTxt = successTxt + yesterday
	}
	return successTxt
}

func fuzzySearchApps(query string, apps *[]App) []string {
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

		// get last 20 apps
		if len(result) > 20 {
			result = result[len(result)-20:]
		}
	}
	return result
}

func getAppByID(id int) *App {
	mutex.RLock()
	defer mutex.RUnlock()
	for i := range apps {
		if apps[i].AppID == id {
			return &apps[i]
		}
	}
	return nil
}

// PHOTO FUNCTIONS

func getCachedPhotoPath(appID int) string {
	cacheDir := "cache"
	os.MkdirAll("cache", os.ModePerm)

	name := fmt.Sprintf("%d.jpg", appID)
	path := filepath.Join(cacheDir, name)

	if _, err := os.Stat(path); err == nil {
		return path
	}

	url := fmt.Sprintf("https://cdn.akamai.steamstatic.com/steam/apps/%d/header.jpg", appID)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return ""
	}
	defer resp.Body.Close()

	file, err := os.Create(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	io.Copy(file, resp.Body) // copy url's jpg to file

	return path
}

// FUNCTIONS FOR WORKING WITH INLINE KEYBOARD

func createSuggestionsKeyboard(suggestions []string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	var allMatches []*App
	seenIDs := make(map[int]bool)
	nameCounts := make(map[string]int)

	for _, name := range suggestions {
		matches, _ := getAppsByName(name) // search for info about an app by its name. by the way, it can return more than 1 app, so matches may be larger than suggestions
		for _, app := range matches {
			if !seenIDs[app.AppID] { // if we haven't processed this game yet
				allMatches = append(allMatches, app)
				seenIDs[app.AppID] = true
				nameCounts[app.Name]++
			}
		}
	}

	for _, app := range allMatches {
		label := app.Name

		// if there is more than one game with the same name - add the id on label
		if nameCounts[app.Name] > 1 {
			label = fmt.Sprintf("%s (ID: %d)", app.Name, app.AppID)
		}

		data := fmt.Sprintf("id_%d", app.AppID)
		btn := tgbotapi.NewInlineKeyboardButtonData(label, data)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func createCheckAgainKeyboard(appID int) tgbotapi.InlineKeyboardMarkup {
	callbackData := fmt.Sprintf("refresh_%d", appID)

	btn := tgbotapi.NewInlineKeyboardButtonData("🔄 Check Again", callbackData)
	return tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(btn))
}

// this is lightweight createSuggestionsKeyboard() because
// this func don't need to process slice of app names to calculate suggestion list.
// it gets matches right away
func createMatchesKeyboard(matches []*App) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, app := range matches {
		data := fmt.Sprintf("id_%d", app.AppID)

		label := fmt.Sprintf("%s (ID: %d)", app.Name, app.AppID)
		btn := tgbotapi.NewInlineKeyboardButtonData(label, data)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// APP LIST FUNCTIONS

func loadApps() *[]App {
	file, err := os.Open("games_appid.json")
	if err != nil {
		return nil
	}
	defer file.Close()

	var apps []App
	if err := json.NewDecoder(file).Decode(&apps); err != nil {
		return &apps
	}
	return &apps
}

func getAppList() {
	url := "https://raw.githubusercontent.com/jsnli/steamappidlist/refs/heads/master/data/games_appid.json"
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("[ERROR] failed to get app list: %v", err)
		return
	}
	defer resp.Body.Close()

	file, err := os.Create("temp.json")
	if err != nil {
		log.Printf("[ERROR] failed to create app list file: %v", err)
		return
	}

	bytesWritten, err := io.Copy(file, resp.Body)
	if err != nil {
		file.Close()
		log.Printf("[ERROR] failed to copy app list stream to file: %v", err)
		return
	}
	file.Close()

	err = os.Rename("temp.json", "games_appid.json")
	if err != nil {
		log.Printf("[ERROR] failed to rename app list file: %v", err)
		return
	}

	loaded := loadApps()
	if loaded != nil {
		mutex.Lock()
		apps = *loaded
		mutex.Unlock()
		log.Printf("[SUCCESS] successfully updated app list. total games: %d. bytes: %d", len(apps), bytesWritten)
	} else {
		log.Printf("[ERROR] failed to load apps from the new JSON file")
	}
}

func setupCron() {
	c := cron.New()

	_, err := c.AddFunc("0 3 * * *", getAppList)
	if err != nil {
		log.Printf("[ERROR] failed to schedule getting fresh app list JSON: %v", err)
	}

	//	_, err = c.AddFunc("0 0 1 * *", func() {
	//		if err := os.RemoveAll("cache"); err != nil {
	//			log.Printf("[ERROR] failed to delete cache: %v", err)
	//			return
	//		}
	//		log.Printf("[SUCCESS] cache cleared successfully")
	//	})
	//	if err != nil {
	//		log.Printf("[ERROR] failed to schedule cache clearing: %v", err)
	//	}

	c.Start()
	log.Print("[SUCCESS] cron started")
}

// SENDING MESSAGE FUNCTIONS

func sendFinalMessage(bot *tgbotapi.BotAPI, chatID int64, app *App, text string) {
	pathToPhoto := getCachedPhotoPath(app.AppID)
	keyboard := createCheckAgainKeyboard(app.AppID)

	var msg tgbotapi.Chattable

	if pathToPhoto != "" {
		photoMsg := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(pathToPhoto))
		photoMsg.Caption = text
		photoMsg.ParseMode = "HTML"
		photoMsg.ReplyMarkup = keyboard
		msg = photoMsg
	} else {
		textMsg := tgbotapi.NewMessage(chatID, text)
		textMsg.ParseMode = "HTML"
		textMsg.ReplyMarkup = keyboard
		msg = textMsg
	}

	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("[ERROR] failed to send final message for app %d: %v", app.AppID, err)
	}
}
