package main

import (
	"bytes"
	"math/big"
)

var b58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

// 将字节数组编码为Base58
func Base58Encode(input []byte) []byte {
	var result []byte
	x := big.NewInt(0).SetBytes(input)
	base := big.NewInt(int64(len(b58Alphabet)))
	zero := big.NewInt(0)
	mod := &big.Int{}
	//比较x和zero的大小，大于为 1，相等为 0
	for x.Cmp(zero) != 0 {
		//如果base!= 0将x设为x/base，将mod设为x%base并返回(x, mod)；如果y == 0会panic。采用欧几里德除法（和Go不同）
		x.DivMod(x, base, mod)
		result = append(result, b58Alphabet[mod.Int64()])
	}
	//0x00版本
	if input[0] == 0x00 {
		result = append(result, b58Alphabet[0])
	}
	//反转字符串
	ReverseBytes(result)
	return result
}

// 对Base58编码的数据进行解码
func Base58Decode(input []byte) []byte {
	result := big.NewInt(0)
	for _, b := range input {
		charIndex := bytes.IndexByte(b58Alphabet, b)
		//相乘
		result.Mul(result, big.NewInt(58))
		//相加
		result.Add(result, big.NewInt(int64(charIndex)))
	}
	decoded := result.Bytes()
	//返回zeroBytes个byte串联形成的新的切片
	if input[0] == b58Alphabet[0] {
		decoded = append([]byte{0x00}, decoded...)
	}

	return decoded
}
