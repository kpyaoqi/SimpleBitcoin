package main

import (
	"bytes"
	"encoding/binary"
	"log"
)

// 将int64转换为字节数组
func IntToHex(i int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, i)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}
