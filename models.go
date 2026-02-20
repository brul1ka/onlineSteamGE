package main

type SteamResponse struct {
	Response struct {
		PlayerCount int `json:"player_count"`
		Result      int `json:"result"`
	} `json:"response"`
}

type AppItem struct {
	AppID             int    `json:"appid"`
	Name              string `json:"name"`
	LastModified      int    `json:"last_modified"`
	PriceChangeNumber int    `json:"price_change_number"`
}

type Apps []AppItem
