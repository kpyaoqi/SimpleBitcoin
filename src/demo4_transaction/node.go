package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	bolt "go.etcd.io/bbolt"
	"log"
	"os"
	"time"
)

// 区块
type Block struct {
	Timestamp     int64          //时间戳
	Transactions  []*Transaction //存储交易
	PrevBlockHash []byte         //前面区块的哈希值
	Hash          []byte         //哈希值
	Nonce         int            //用于找到pow
}

// 保存区块链
type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

const dbFile = "blockchain.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "Yaoqi's Blockchain"

// 创建区块
func NewBlock(transactions []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), transactions, prevBlockHash, []byte{}, 0}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

// 添加区块
func (bc *Blockchain) MineBlock(transactions []*Transaction) {
	var lastHash []byte
	//获取最后一个块的哈希
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	//挖出一个新的块
	newBlock := NewBlock(transactions, lastHash)
	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		b.Put(newBlock.Hash, newBlock.Serialize())
		b.Put([]byte("l"), newBlock.Hash)
		bc.tip = newBlock.Hash
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

// 设置创世区块
func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

// 检查链是否已经存在
func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

// 创建一个新的区块链
func NewBlockchain(address string) *Blockchain {
	if dbExists() == false {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}
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
		tip = b.Get([]byte("l"))
		return nil
	})
	//创建 Blockchain
	bc := Blockchain{tip, db}
	return &bc
}

// 创建一个新的区块链数据库
func CreateBlockchain(address string) *Blockchain {
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}
	var tip []byte
	//打开一个 BoltDB 文件
	db, _ := bolt.Open(dbFile, 0600, nil)
	db.Update(func(tx *bolt.Tx) error {
		cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
		genesis := NewGenesisBlock(cbtx)
		b, _ := tx.CreateBucket([]byte(blocksBucket))
		b.Put(genesis.Hash, genesis.Serialize())
		b.Put([]byte("l"), genesis.Hash)
		tip = genesis.Hash
		return nil
	})
	bc := Blockchain{tip, db}
	return &bc
}

// 计算区块里所有交易的哈希
func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte
	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))
	return txHash[:]
}

func (bc *Blockchain) FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput
	unspentTransactions := bc.FindUnspentTransactions(address)
	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
				UTXOs = append(UTXOs, out)
			}

		}
	}
	return UTXOs
}
