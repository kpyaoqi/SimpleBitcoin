package main

import (
	"bytes"
	"crypto/sha256"
	"strconv"
	"time"
)

// 区块
type Block struct {
	Timestamp     int64  //时间戳
	Data          []byte //存储交易
	PrevBlockHash []byte //前面区块的哈希值
	Hash          []byte //哈希值
	Nonce         int    //用于找到pow
}

// 保存区块链
type Blockchain struct {
	blocks []*Block
}

const dbFile = "blockchain.db"
const blocksBucket = "blocks"

// 设置哈希
func (b *Block) SetHash() {
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})
	hash := sha256.Sum256(headers)
	b.Hash = hash[:]
}

// 创建区块
func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}, 0}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

// 添加区块
func (bc *Blockchain) AddBlock(data string) {
	prevBlock := bc.blocks[len(bc.blocks)-1]
	newBlock := NewBlock(data, prevBlock.Hash)
	bc.blocks = append(bc.blocks, newBlock)
}

// 设置创世区块
func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}

// 创建一个新的区块链
func NewBlockchain() *Blockchain {
	return &Blockchain{[]*Block{NewGenesisBlock()}}
}
