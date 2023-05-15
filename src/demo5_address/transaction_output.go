package main

import "bytes"

// TXOutput 包含两部分
// Value: 有多少币，就是存储在 Value 里面
// PubKeyHash: 锁定脚本
type TXOutput struct {
	Value      int
	PubKeyHash []byte
}

// 锁定一个输出
func (out *TXOutput) Lock(address []byte) {
	pubKeyHash := Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

// 检查是否提供的公钥哈希被用于锁定输出
func (out *TXOutput) IsLockedWithKey(pubkeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubkeyHash) == 0
}

// 创建一个新的输出交易
func NewTXOutput(value int, address string) *TXOutput {
	txo := &TXOutput{value, nil}
	txo.Lock([]byte(address))
	return txo
}
