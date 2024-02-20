package models

type OrderHistoryQueryAll struct {
	ID                  string `json:"id"`
	OrderType           string `json:"orderType"`
	Symbol              string `json:"symbol"`
	Side                string `json:"side"`
	Price               string `json:"price"`
	TriggerPrice        string `json:"triggerPrice"`
	Quantity            string `json:"quantity"`
	QuoteQuantity       string `json:"quoteQuantity"`
	TimeInForce         string `json:"timeInForce"`
	SelfTradePrevention string `json:"selfTradePrevention"`
	PostOnly            bool   `json:"postOnly"`
	Status              string `json:"status"`
}
