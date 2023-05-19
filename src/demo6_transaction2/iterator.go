package main

import (
	bolt "go.etcd.io/bbolt"
	"log"
)

type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

// 返回链中的下一个块
func (bi *BlockchainIterator) Next() *Block {
	var block *Block
	err := bi.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(bi.currentHash)
		block = DeserializeBlock(encodedBlock)
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	bi.currentHash = block.PrevBlockHash
	return block
}
