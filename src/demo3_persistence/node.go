package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	bolt "go.etcd.io/bbolt"
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
	tip []byte
	db  *bolt.DB
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
	var lastHash []byte
	//获取最后一个块的哈希
	err := bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		return nil
	})
	if err != nil {
		fmt.Println("Update fail")
	}
	//挖出一个新的块
	newBlock := NewBlock(data, lastHash)
	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		b.Put(newBlock.Hash, newBlock.Serialize())
		b.Put([]byte("l"), newBlock.Hash)
		bc.tip = newBlock.Hash
		return nil
	})
	if err != nil {
		fmt.Println("Update fail")
	}
}

// 设置创世区块
func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}

// 创建一个新的区块链
func NewBlockchain() *Blockchain {
	var tip []byte
	//打开一个 BoltDB 文件
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		fmt.Println("Open fail")
		return nil
	}
	//打开一个读写事务
	db.Update(func(tx *bolt.Tx) error {
		//获取了存储区块的 bucket
		b := tx.Bucket([]byte(blocksBucket))
		//如果不存在，就生成创世块，创建 bucket
		if b == nil {
			genesis := NewGenesisBlock()
			b, err := tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				fmt.Println("CreateBucket fail")
				return nil
			}
			err = b.Put(genesis.Hash, genesis.Serialize())
			err = b.Put([]byte("l"), genesis.Hash)
			tip = genesis.Hash
		} else {
			//如果存在，就从中读取 l 键
			tip = b.Get([]byte("l"))
		}
		return nil
	})
	//创建 Blockchain
	bc := Blockchain{tip, db}
	return &bc
}
