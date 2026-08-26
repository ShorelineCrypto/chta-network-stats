package models

import "time"

type NetworkBlock struct {
	Height       int64
	Hash         string
	Difficulty   float64
	Bits         string
	Time         time.Time
	Gap          int64
	TxCount      int
	Type         string
	CoinbaseTxID string
	CoinbaseHex  string
	MinerTag     string `json:"miner_tag"`
}
