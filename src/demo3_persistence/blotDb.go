package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

// 对这些结构进行序列化
func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(b)
	if err != nil {
		fmt.Println("Encode file")
		return nil
	}
	return result.Bytes()
}

// 解序列化
func DeserializeBlock(d []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		fmt.Println("Decode file")
		return nil
	}
	return &block
}
