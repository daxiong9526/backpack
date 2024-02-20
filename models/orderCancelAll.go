package models

type OrderCancelAll struct {
	OrderType             string `json:"orderType"`
	ID                    string `json:"id"`
	ClientID              int64  `json:"clientId"` // 假设客户端ID是整数
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
	CreatedAt             int64  `json:"createdAt"` // 假设创建时间是以某种形式的整数时间戳表示
}
