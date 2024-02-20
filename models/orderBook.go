package models

type OrderBook struct {
	Asks         [][]string `json:"asks"`
	Bids         [][]string `json:"bids"`
	LastUpdateId string     `json:"lastUpdateId"`
}
