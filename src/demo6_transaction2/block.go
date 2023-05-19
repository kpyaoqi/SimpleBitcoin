package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
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

// 设置创世区块
func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

// 创建区块
func NewBlock(transactions []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), transactions, prevBlockHash, []byte{}, 0}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce
	return block
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

// 对Block结构进行序列化
func (b *Block) Serialize() []byte {
	//Buffer是一个实现了读写方法的可变大小的字节缓冲
	var result bytes.Buffer
	//NewEncoder返回一个将编码后数据写入result的*Encoder
	encoder := gob.NewEncoder(&result)
	//Encode方法将b编码后发送，并且会保证所有的类型信息都先发送
	err := encoder.Encode(b)
	if err != nil {
		fmt.Println("Encode file")
		return nil
	}
	//返回未读取部分字节数据的切片
	return result.Bytes()
}

// 对Block结构进行解序列化
func DeserializeBlock(d []byte) *Block {
	var block Block
	//函数返回一个从r读取数据的*Decoder，如果r不满足io.ByteReader接口，则会包装r为bufio.Reader。
	decoder := gob.NewDecoder(bytes.NewReader(d))
	//Decode从输入流读取下一个之并将该值存入&block。如果e是nil，将丢弃该值；否则e必须是可接收该值的类型的指针。如果输入结束，方法会返回io.EOF并且不修改e（指向的值）
	err := decoder.Decode(&block)
	if err != nil {
		fmt.Println("Decode file")
		return nil
	}
	return &block
}
