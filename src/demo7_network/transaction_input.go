package main

import "bytes"

// TXInput 包含 3 部分
// Txid: 一个交易输入引用了之前一笔交易的一个输出, ID表明是之前哪笔交易
// Vout: 一笔交易可能有多个输出，Vout 为输出的索引
// Signature: 签名数据
// PubKey: 公钥
type TXInput struct {
	Txid      []byte
	Vout      int
	Signature []byte
	PubKey    []byte
}

// 检查输入使用了指定密钥来解锁一个输出
func (in *TXInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := HashPubKey(in.PubKey)
	return bytes.Compare(lockingHash, pubKeyHash) == 0
}
