package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"golang.org/x/crypto/ripemd160"
	"log"
)

const version_w = byte(0x00)
const addressChecksumLen = 4

// ecdsa.PrivateKey代表一个ECDSA私钥
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

// 创建并返回一个钱包
func NewWallet() *Wallet {
	private, public := newKeyPair()
	wallet := Wallet{private, public}
	return &wallet
}

func newKeyPair() (ecdsa.PrivateKey, []byte) {
	//返回一个实现了P-256的曲线
	curve := elliptic.P256()
	//GenerateKey函数生成一对
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
	return *private, pubKey
}

// 获取钱包地址
func (w Wallet) GetAddress() []byte {
	//使用 RIPEMD160(SHA256(PubKey)) 哈希算法
	pubKeyHash := HashPubKey(w.PublicKey)
	//给哈希加上地址生成算法版本的前缀
	versionedPayload := append([]byte{version_w}, pubKeyHash...)
	//计算校验和
	checksum := checksum(versionedPayload)
	fullPayload := append(versionedPayload, checksum...)
	//使用 Base58 对组合进行编码
	address := Base58Encode(fullPayload)
	return address
}

// 对公钥取哈希
func HashPubKey(pubKey []byte) []byte {
	publicSHA256 := sha256.Sum256(pubKey)
	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSHA256[:])
	if err != nil {
		log.Panic(err)
	}
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)
	return publicRIPEMD160
}

// 计算校验和(校验和是结果哈希的前四个字节)
func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])
	return secondSHA[:addressChecksumLen]
}

// 检查地址
func ValidateAddress(address string) bool {
	pubKeyHash := Base58Decode([]byte(address))
	actualChecksum := pubKeyHash[len(pubKeyHash)-addressChecksumLen:]
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-addressChecksumLen]
	targetChecksum := checksum(append([]byte{version}, pubKeyHash...))
	return bytes.Compare(actualChecksum, targetChecksum) == 0
}
