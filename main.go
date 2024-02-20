package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v2"

	"backpack/client"
)

var apiURL = "https://api.backpack.exchange/api/v1/"

// Config 结构用于加载配置文件
type Config struct {
	ApiKey     string  `yaml:"api_key"`
	ApiSecret  string  `yaml:"api_secret"`
	OrderType  string  `yaml:"order_type"`
	Counts     int64   `yaml:"counts"`
	Total      float64 `yaml:"total"`
	WaitSecond int     `yaml:"wait_second"`
}

// 加载配置文件
func (c *Config) loadConfig(configPath string) error {
	yamlFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		return err
	}
	return nil
}

// 生成私钥
func (c *Config) generPrivateKey() ed25519.PrivateKey {
	// 从Base64编码解码种子和公钥
	seed, err := base64.StdEncoding.DecodeString(c.ApiSecret)
	if err != nil {
		fmt.Println("API SECRET 错误:", err)
		return nil
	}

	publicKey, err := base64.StdEncoding.DecodeString(c.ApiKey)
	if err != nil {
		fmt.Println("API KEY 错误:", err)
		return nil
	}

	// 检查种子和公钥长度是否正确
	if len(seed) != ed25519.SeedSize || len(publicKey) != ed25519.PublicKeySize {
		fmt.Println("API SECRET 或者 API KEY 长度错误")
		return nil
	}

	// 将种子和公钥拼接成私钥
	privateKey := ed25519.NewKeyFromSeed(seed)
	// 从私钥对象获取公钥
	gpublicKey := privateKey.Public().(ed25519.PublicKey)
	if string(publicKey) == string(gpublicKey) {
		fmt.Println("API KEY 验证通过.")
	} else {
		fmt.Println("API KEY 验证不通过，请确认填写是否正确.")
		os.Exit(0)
	}

	return privateKey
}

// 生成随机数
func creatRandNum(max int) int {
	rand.Seed(time.Now().UnixNano())
	// 生成0到20之间的随机数
	randomNumber := rand.Intn(max)
	return randomNumber
}

/*
	策略 ：
	1、查询未成交的订单
	2、如果有未成交的订单，等1s以后继续查询
	3、如果几秒后仍然未成交，那么取消
	4、检查余额
	5、开单：如果上一笔是卖，那么开买单；如果上一笔是买，那么开卖单
	6、结束条件：
		a. 余额跑完
		b. 次数跑完
		c. 金额跑完
*/
func strategy(api *client.Api, config *Config) {
	lastOrderIsBid := false
	lastOrderClientId := int64(0)
	lastOrderId := ""
	for {
		if config.Counts > 0 && config.Counts <= lastOrderClientId {
			// 总次数不足,策略结束
			fmt.Println("总次数跑完,策略结束")
			return
		}
		fmt.Println("                                    ")
		fmt.Println("====================================")
		fmt.Println("                                    ")
		fmt.Printf("第%d次交易开始\n", lastOrderClientId+1)
		if lastOrderClientId > 0 && lastOrderId != "" {
			// 生成随机检测总次数
			tmpTotalCount := creatRandNum(10)
			fmt.Printf("随机查询未成交的订单%d次\n", tmpTotalCount)
			count := 0
			for {
				// 查询未成交的订单
				openOrder := api.GetOrder(lastOrderClientId)
				fmt.Printf("第%d次查询未成交的订单\n", count+1)
				// 如果有未成交的订单，等1s以后继续查询
				if openOrder.ClientId > 0 {
					time.Sleep(1 * time.Second)
					count++
					// 如果几秒后仍然未成交，那么取消
					if count >= tmpTotalCount {
						api.CancelOrders()
						fmt.Println("取消未成交的订单")
						break
					}
				} else {
					break
				}
			}
		}

		totalSize := caluTotalSize(api)
		fmt.Printf("已经完成交易量总和：%f， 估算总磨损手续费：%f \n", totalSize, totalSize*0.00085)
		if config.Total > 0 && config.Total <= totalSize {
			// 刷到足够量,策略结束
			fmt.Println("刷到足够量,策略结束")
			return
		}

		// 获取价格
		api.GetPrice()
		if api.Price.Ask == 0 || api.Price.Bid == 0 {
			fmt.Println("获取价格失败")
			continue
		}
		// 检查余额
		api.GetBalance()
		if api.Balances.SOL*api.Price.Ask <= 1 && api.Balances.USDC <= 1 {
			// 余额不足,策略结束
			fmt.Println("余额不足,策略结束")
			return
		}

		// 随机开买单还是卖单
		if api.Balances.SOL*api.Price.Ask > 1 && api.Balances.USDC > 1 {
			tmpNum := creatRandNum(2)
			tmpNum2 := tmpNum % 2
			if tmpNum2 == 0 {
				lastOrderIsBid = true
			} else {
				lastOrderIsBid = false
			}
		} else if api.Balances.USDC <= 1 {
			lastOrderIsBid = true
		} else {
			lastOrderIsBid = false
		}
		// 开单：如果上一笔是卖，那么开买单；如果上一笔是买，那么开卖单
		// 市价：Market，卖一价*1.05买，买一价*0.95卖
		// 限价：Limit，卖一价/买一价下单
		lastOrderClientId++
		if lastOrderIsBid == false {
			// 开买单
			price := api.Price.Ask
			if config.OrderType == "Market" {
				price = api.Price.Ask * 1.02
			} else {
				price = api.Price.Bid
			}
			quantity := api.Balances.USDC / price
			quantity = math.Floor(quantity*100) / 100
			price = math.Floor(price*100) / 100
			lastOrderId = api.OpenOrder(lastOrderClientId, config.OrderType, price, quantity, "Bid")
			lastOrderIsBid = true
		} else {
			// 开卖单，市价卖防止持有sol价格跌造成磨损
			price := api.Price.Bid * 0.999
			price = math.Floor(price*100) / 100
			quantity := api.Balances.SOL
			quantity = math.Floor(quantity*100) / 100
			lastOrderId = api.OpenOrder(lastOrderClientId, "Market", price, quantity, "Ask")
			lastOrderIsBid = false
		}

		// 随机等待几秒
		if config.WaitSecond == 0 {
			config.WaitSecond = 1
		}
		tmpWaitSecond := creatRandNum(config.WaitSecond)
		if tmpWaitSecond == 0 {
			tmpWaitSecond = 1
		}
		fmt.Printf("随机等待%ds....\n", tmpWaitSecond)
		time.Sleep(time.Duration(tmpWaitSecond) * time.Second)
	}
}

func caluTotalSize(api *client.Api) float64 {
	total := float64(0)
	page := int64(1)
	limit := int64(1000)
	for {
		orderHistory := api.OrderHistory(page, limit)
		if len(orderHistory) < 1 {
			return total
		}
		for _, order := range orderHistory {
			if order.Status != "Filled" {
				continue
			}
			price, err := strconv.ParseFloat(order.Price, 64)
			if err != nil {
				fmt.Printf("Error converting Price to float64: %v\n", err)
				continue
			}

			quantity, err := strconv.ParseFloat(order.Quantity, 64)
			if err != nil {
				fmt.Printf("Error converting Quantity to float64: %v\n", err)
				continue
			}

			total += price * quantity
		}
		page = page + limit
	}
}

func main() {
	// 加载配置
	var config Config
	err := config.loadConfig("config.yaml")
	if err != nil {
		log.Fatalf("  配置文件读取失败: %v", err)
	}

	var api client.Api
	api.PrivateKey = config.generPrivateKey()
	api.ApiKey = config.ApiKey
	api.Symbol = "SOL_USDC"

	strategy(&api, &config)
}
