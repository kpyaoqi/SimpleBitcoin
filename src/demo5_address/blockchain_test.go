package main

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

func TestBase58Encode(t *testing.T) {
	block := &Block{time.Now().Unix(), nil, []byte{}, []byte{}, 0}
	serialize := block.Serialize()
	fmt.Println(serialize)
}

func TestByte(t *testing.T) {
	var b58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")
	for i, b := range b58Alphabet {
		fmt.Printf("%d:%x\n", i, b)
	}
}

func TestAddress(t *testing.T) {
	address := ValidateAddress("1B35qgq45pA5mpxFokzSA2nYEBAVTUWRgi")
	fmt.Sprintf(strconv.FormatBool(address))
}
