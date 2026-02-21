package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

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

func searchAppByName(name string) (*AppItem, []string) {
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
		return "Game with this appID not found!"
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

func handleAppRequest(name string) (string, []string) {
	app, suggestions := searchAppByName(name)

	if app != nil {
		return getAppOnline(app), nil
	}
	if len(suggestions) > 0 {
		if len(suggestions) == 1 {
			autoGame, _ := searchAppByName(suggestions[0])
			return getAppOnline(autoGame), nil
		}
		return "Maybe you meant:", suggestions
	}
	return "There is no such game with this name!", nil
}
