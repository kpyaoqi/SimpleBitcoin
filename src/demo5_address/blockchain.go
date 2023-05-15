package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	bolt "go.etcd.io/bbolt"
	"log"
	"os"
)

const dbFile = "blockchain.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "Yaoqi's Blockchain"

// 保存区块链
type Blockchain struct {
	//数据库中存储的最后一个块的哈希
	tip []byte
	db  *bolt.DB
}

// 创建一个新的区块链数据库(address为创世区块的出块奖励地址)
func CreateBlockchain(address string) *Blockchain {
	//判断是否已经存在
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}
	var tip []byte
	//打开一个 BoltDB 文件
	db, _ := bolt.Open(dbFile, 0600, nil)
	db.Update(func(tx *bolt.Tx) error {
		//构建coinbase交易
		cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
		genesis := NewGenesisBlock(cbtx)
		//创建blocks的bucket
		b, _ := tx.CreateBucket([]byte(blocksBucket))
		//存入创世区块(键为创世区块的hash)
		b.Put(genesis.Hash, genesis.Serialize())
		//存入键为“l”的表示为最后一个区块的hash
		b.Put([]byte("l"), genesis.Hash)
		tip = genesis.Hash
		return nil
	})
	bc := Blockchain{tip, db}
	return &bc
}

// 检查链是否已经存在
func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

// 添加区块
func (bc *Blockchain) MineBlock(transactions []*Transaction) {
	var lastHash []byte
	//在一笔交易被放入一个块之前进行验证
	for _, tx := range transactions {
		if bc.VerifyTransaction(tx) != true {
			log.Panic("ERROR:Invalid transaction")
		}
	}
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

// 定位到指定区块链
func PositioningBlockchain() *Blockchain {
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

// 通过交易ID查找交易
func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator()
	for {
		block := bci.Next()
		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return Transaction{}, errors.New("Transaction is not found")
}

// 查找并返回所有未使用的交易输出
func (bc *Blockchain) FindUTXO(pubKeyHash []byte) []TXOutput {
	var UTXOs []TXOutput
	unspentTransactions := bc.FindUnspentTransactions(pubKeyHash)
	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}
	return UTXOs
}

// 从 address 中找到至少 amount 的 UTXO
func (bc *Blockchain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspenfTXs := bc.FindUnspentTransactions(pubKeyHash)
	accumulated := 0
Work:
	for _, tx := range unspenfTXs {
		txID := hex.EncodeToString(tx.ID)
		for outIdx, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				if accumulated >= amount {
					break Work
				}
			}
		}
	}
	return accumulated, unspentOutputs
}

// 找到某个地址所有的未花费输出
func (bc *Blockchain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
	var unspentTXs []Transaction
	spentUTXOs := make(map[string][]int)
	bci := bc.Iterator()
	for {
		block := bci.Next()
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)
		Outputs:
			for outIdx, out := range tx.Vout {
				//如果交易输出被花费了
				if spentUTXOs[txID] != nil {
					for _, spendOut := range spentUTXOs[txID] {
						if spendOut == outIdx {
							continue Outputs
						}
					}
				}
				//如果该交易输出可以被解锁，即可被花费
				if out.IsLockedWithKey(pubKeyHash) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}
			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					if in.UsesKey(pubKeyHash) {
						inTxID := hex.EncodeToString(in.Txid)
						spentUTXOs[inTxID] = append(spentUTXOs[inTxID], in.Vout)
					}
				}
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return unspentTXs
}

// 对交易进行签名
func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)
	//对交易的每一笔输入交易进行签名
	for _, vin := range tx.Vin {
		//找到输入交易来自哪个交易
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		//存储在prevTXs中
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}
	tx.Sign(privKey, prevTXs)
}

// 验证交易输入签名
func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	prevTXs := make(map[string]Transaction)
	for _, vin := range tx.Vin {
		prevTX, _ := bc.FindTransaction(vin.Txid)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}
	return tx.Verify(prevTXs)

}

// 返回一个BlockchainIterat
func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.db}
	return bci
}
