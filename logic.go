package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func getAppByName(name string) (*AppItem, error) {
	file, _ := os.Open("games_appid.json")
	defer file.Close()

	var apps Apps
	if err := json.NewDecoder(file).Decode(&apps); err != nil {
		return nil, err
	}

	for _, app := range apps {
		if strings.EqualFold(app.Name, name) {
			return &app, nil
		}
	}
	return nil, fmt.Errorf("game not found")
}

func GetSteamOnline(app *AppItem) string {
	if app == nil {
		return "There is no such game with this name!"
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
