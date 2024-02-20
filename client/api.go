package client

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"backpack/models"
)

var apiURL = "https://api.backpack.exchange/api/v1/"
var apiURLW = "https://api.backpack.exchange/wapi/v1/"

type Balances struct {
	SOL  float64
	USDC float64
}

type Price struct {
	Bid    float64
	BidVol float64
	Ask    float64
	AskVol float64
}

type Api struct {
	ApiKey     string
	PrivateKey ed25519.PrivateKey
	Symbol     string
	Balances   Balances
	Price      Price
}

// 生成签名
func (a *Api) generateSignature(instruction string, params map[string]interface{}, timestamp int64, window int) string {
	// 解码私钥
	privKey := a.PrivateKey

	// 准备参数
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var paramString strings.Builder
	paramString.WriteString("instruction=" + instruction)
	for _, k := range keys {
		paramString.WriteString(fmt.Sprintf("&%s=%v", k, params[k]))
	}
	paramString.WriteString(fmt.Sprintf("&timestamp=%d&window=%d", timestamp, window))

	// 签名
	signature := ed25519.Sign(privKey, []byte(paramString.String()))
	return base64.StdEncoding.EncodeToString(signature)
}

// PrettyPrintRequest 以人类可读的格式打印HTTP请求的关键信息
func (a *Api) prettyPrintRequest(req *http.Request) {
	var reqText bytes.Buffer

	// 打印请求方法和URL
	reqText.WriteString(fmt.Sprintf("Method: %s\nURL: %s\n", req.Method, req.URL.String()))

	// 打印查询参数（如果有）
	if query := req.URL.Query(); len(query) > 0 {
		reqText.WriteString("Query Parameters:\n")
		for key, value := range query {
			reqText.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
	}

	// 打印请求头部（如果有）
	if len(req.Header) > 0 {
		reqText.WriteString("Headers:\n")
		for key, value := range req.Header {
			reqText.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
		}
	}

	// 打印请求体（如果适用且可用）
	if req.Body != nil && req.ContentLength > 0 {
		reqText.WriteString("Body: \n")
		reqText.WriteString(fmt.Sprintf("  %s\n", req.Body))
	}

	fmt.Print(reqText.String())
}

// 构建http请求
func (a *Api) http(method string, urlStr string, instruction string, params map[string]interface{}) (int, string) {

	if method == "GET" {
		// 对于GET请求，将参数添加到URL的查询字符串中
		paramsValues := url.Values{}
		for key, value := range params {
			paramsValues.Add(key, value.(string))
		}
		if len(params) > 0 {
			urlStr += "?" + paramsValues.Encode()
		}
	}
	// 对于POST请求，将参数编码为JSON或表单数据
	jsonData, err := json.Marshal(params)
	if err != nil {
		fmt.Printf("Error converting map to JSON: %v", err)
		return 0, ""
	}

	// 构建HTTP请求
	client := &http.Client{}
	req, err := http.NewRequest(method, urlStr, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return 0, ""
	}

	timestamp := time.Now().UnixMilli()
	window := 500000

	signature := a.generateSignature(instruction, params, timestamp, window)

	// 添加必要的头参数
	req.Header.Add("X-API-KEY", a.ApiKey)
	req.Header.Add("X-SIGNATURE", signature)
	req.Header.Add("X-TIMESTAMP", strconv.FormatInt(timestamp, 10))
	req.Header.Add("X-WINDOW", strconv.Itoa(window))
	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return 0, ""
	}
	defer resp.Body.Close()

	// 读取并打印响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return 0, ""
	}
	return resp.StatusCode, string(body)
}

// 查询余额
func (a *Api) GetBalance() {
	instruction := "balanceQuery"
	params := map[string]interface{}{}
	statusCode, response := a.http("GET", apiURL+"capital", instruction, params)
	if statusCode != 200 {
		fmt.Println("获取余额失败")
		return
	}

	// 将要解析到的结构体实例
	var balances models.Balances

	// 解析JSON字符串到结构体
	err := json.Unmarshal([]byte(response), &balances)
	if err != nil {
		fmt.Println("结果解析失败:", response)
		return
	}
	a.Balances.SOL = balances.SOL.Available
	a.Balances.USDC = balances.USDC.Available

	// 遍历Balances结构体
	fmt.Printf("当前余额:  SOL : %f， USDC : %f\n", a.Balances.SOL, a.Balances.USDC)
}

// 查询订单
func (a *Api) GetOrder(clientId int64) models.OrderQuery {
	instruction := "orderQuery"
	params := map[string]interface{}{
		"symbol":   a.Symbol,
		"clientId": strconv.FormatInt(clientId, 10),
	}
	statusCode, response := a.http("GET", apiURL+"order", instruction, params)
	if statusCode == 404 {
		return models.OrderQuery{}
	}

	// 将要解析到的结构体实例
	var order models.OrderQuery

	// 解析JSON字符串到结构体
	err := json.Unmarshal([]byte(response), &order)
	if err != nil {
		fmt.Println("结果解析失败:", response)
		return models.OrderQuery{}
	}
	return order
}

// 取消所有未成交订单
func (a *Api) CancelOrders() bool {
	instruction := "orderCancelAll"
	params := map[string]interface{}{
		"symbol": a.Symbol,
	}
	_, response := a.http("DELETE", apiURL+"orders", instruction, params)
	var orders []models.OrderCancelAll
	err := json.Unmarshal([]byte(response), &orders)
	if err != nil {
		fmt.Println("结果解析失败:", response)
	}

	return true
}

// 获取sol的市场价格
func (a *Api) GetPrice() {
	instruction := ""
	params := map[string]interface{}{
		"symbol": a.Symbol,
	}
	_, response := a.http("GET", apiURL+"depth", instruction, params)
	var orderBook models.OrderBook
	err := json.Unmarshal([]byte(response), &orderBook)
	if err != nil {
		fmt.Println("结果解析失败:", response)
		return
	}

	a.Price.Ask, _ = strconv.ParseFloat(orderBook.Asks[0][0], 64)
	a.Price.AskVol, _ = strconv.ParseFloat(orderBook.Asks[0][1], 64)
	a.Price.Bid, _ = strconv.ParseFloat(orderBook.Bids[len(orderBook.Bids)-1][0], 64)
	a.Price.BidVol, _ = strconv.ParseFloat(orderBook.Bids[len(orderBook.Bids)-1][1], 64)
	fmt.Printf("%s 买一价：%f, 卖一价: %f \n", a.Symbol, a.Price.Bid, a.Price.Ask)
}

// 开单
func (a *Api) OpenOrder(clientId int64, orderType string, price float64, quantity float64, side string) string {
	instruction := "orderExecute"
	postOnly := true
	if orderType == "Market" {
		postOnly = false
	}
	params := map[string]interface{}{
		"clientId":            clientId,
		"orderType":           "Limit",
		"postOnly":            postOnly,
		"price":               price,
		"quantity":            quantity,
		"selfTradePrevention": "Allow",
		"side":                side,
		"symbol":              a.Symbol,
		"timeInForce":         "GTC",
	}
	sideStr := ""
	if side == "Bid" {
		sideStr = "买"
	} else {
		sideStr = "卖"
	}
	fmt.Printf("开%s单，交易对：%s，价格：%f，数量：%f \n", sideStr, a.Symbol, price, quantity)
	_, response := a.http("POST", apiURL+"order", instruction, params)
	var orderExecute models.OrderExecute
	err := json.Unmarshal([]byte(response), &orderExecute)
	if err != nil {
		fmt.Println("结果解析失败:", response)
	}
	return orderExecute.Id
}

// 获取所有订单
func (a *Api) OrderHistory(offset int64, limit int64) []models.OrderHistoryQueryAll {
	instruction := "orderHistoryQueryAll"
	params := map[string]interface{}{
		"symbol": a.Symbol,
		"offset": strconv.FormatInt(offset, 10),
		"limit":  strconv.FormatInt(limit, 10),
	}
	_, response := a.http("GET", apiURLW+"history/orders", instruction, params)
	var orders []models.OrderHistoryQueryAll
	err := json.Unmarshal([]byte(response), &orders)
	if err != nil {
		fmt.Println("结果解析失败:", response)
	}
	return orders
}
