package main

type SteamResponse struct {
	Response struct {
		PlayerCount int `json:"player_count"`
		Result      int `json:"result"`
	} `json:"response"`
}

type App struct {
	AppID             int    `json:"appid"`
	Name              string `json:"name"`
	LastModified      int    `json:"last_modified"`
	PriceChangeNumber int    `json:"price_change_number"`
}

type AppDetails struct {
	Success bool `json:"success"`
	Data    struct {
		ReleaseDate struct {
			ComingSoon bool   `json:"coming_soon"`
			Date       string `json:"date"`
		} `json:"release_date"`
	} `json:"data"`
}
