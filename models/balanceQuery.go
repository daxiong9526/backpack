package models

import (
    "encoding/json"
    "strconv"
)

// BalanceDetails 描述了一个货币的余额细节
type BalanceDetails struct {
	Available float64
	Locked    float64
	Staked    float64
}

// Balances 包含了所有货币的余额信息
type Balances struct {
	SOL BalanceDetails `json:"SOL"`
	USDC BalanceDetails `json:"USDC"`
}

// 自定义UnmarshalJSON方法来处理字符串到float64的转换
func (bd *BalanceDetails) UnmarshalJSON(data []byte) error {
    var v map[string]string
    if err := json.Unmarshal(data, &v); err != nil {
        return err
    }
    var err error
    bd.Available, err = strconv.ParseFloat(v["available"], 64)
    if err != nil {
        return err
    }
    bd.Locked, err = strconv.ParseFloat(v["locked"], 64)
    if err != nil {
        return err
    }
    bd.Staked, err = strconv.ParseFloat(v["staked"], 64)
    if err != nil {
        return err
    }
    return nil
}