package main

type SteamResponse struct {
	Response struct {
		PlayerCount int `json:"player_count"`
		Result      int `json:"result"`
	} `json:"response"`
}

type App struct {
	AppID         int    `json:"appID"`
	Name          string `json:"name"`
	IsFree        bool   `json:"is_free"`
	PriceOverview *struct {
		FinalFormatted   string `json:"final_formatted"`
		InitialFormatted string `json:"initial_formatted"`
		DiscountPercent  int    `json:"discount_percent"`
		Final            int    `json:"final"`
		Initial          int    `json:"initial"`
		Currency         string `json:"currency"`
	} `json:"price_overview"`
}
