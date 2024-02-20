package models

// Balances 包含了所有货币的余额信息
type OrderQuery struct {
	OrderType             string `json:"orderType"`
	Id                    string `json:"id"`
	ClientId              uint64 `json:"clientId"`
	Symbol                string `json:"symbol"`
	Side                  string `json:"side"`
	Quantity              string `json:"quantity"`
	ExecutedQuantity      string `json:"executedQuantity"`
	QuoteQuantity         string `json:"quoteQuantity"`
	ExecutedQuoteQuantity string `json:"executedQuoteQuantity"`
	TriggerPrice          string `json:"triggerPrice"`
	TimeInForce           string `json:"timeInForce"`
	SelfTradePrevention   string `json:"selfTradePrevention"`
	Status                string `json:"status"`
	CreatedAt             uint64 `json:"createdAt"`
}
