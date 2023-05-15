package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
)

// coinbase奖励
const subsidy = 10

// Transaction 由交易 ID，输入和输出构成
type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

// 创建一笔新的交易
func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput
	//查找钱包
	wallets, _ := NewWallets()
	wallet := wallets.GetWallet(from)
	//获取from地址的公钥的hash
	pubKeyHash := HashPubKey(wallet.PublicKey)
	//找到至少 amount 的 UTXO
	acc, vaildOutputs := bc.FindSpendableOutputs(pubKeyHash, amount)
	//不足够支付
	if acc < amount {
		log.Panic("ERROR:Not enough funds")
	}
	//足够支付(遍历含有from地址输出的交易)
	for txid, outs := range vaildOutputs {
		txID, _ := hex.DecodeString(txid)
		for _, out := range outs {
			//存入到这笔交易的输入里
			input := TXInput{txID, out, nil, wallet.PublicKey}
			inputs = append(inputs, input)
		}
	}
	outputs = append(outputs, *NewTXOutput(amount, to))
	if acc > amount {
		outputs = append(outputs, *NewTXOutput(acc-amount, from))
	}
	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()
	//签名交易
	bc.SignTransaction(&tx, wallet.PrivateKey)
	return &tx
}

// 签名交易(接受一个私钥和一个之前交易的 map)
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbase() {
		return
	}
	txCopy := tx.TrimmedCopy()
	//迭代副本交易中每一个输入
	for inID, vin := range txCopy.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
		txCopy.ID = txCopy.Hash()
		//在获取完哈希，我们应该重置 PubKey 字段，以便于它不会影响后面的迭代。
		txCopy.Vin[inID].PubKey = nil
		//使用私钥对任意长度的hash值（必须是较大信息的hash结果）进行签名，返回签名结果（一对大整数）
		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID)
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)
		tx.Vin[inID].Signature = signature
	}
}

// 验证函数
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}
	txCopy := tx.TrimmedCopy()
	//返回一个实现了P-256的曲线
	curve := elliptic.P256()
	for inID, vin := range tx.Vin {
		//检查每个输入的签名
		prevTX := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTX.Vout[vin.Vout].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inID].PubKey = nil
		//这个部分跟 Sign 方法一模一样，因为在验证阶段，我们需要的是与签名相同的数据。
		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])
		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PubKey)
		x.SetBytes(vin.PubKey[:(keyLen / 2)])
		y.SetBytes(vin.PubKey[(keyLen / 2):])
		//使用公钥验证hash值和两个大整数r、s构成的签名，并返回签名是否合法。
		rawPubKey := ecdsa.PublicKey{curve, &x, &y}
		if ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) == false {
			return false
		}
	}
	return true
}

// 构建 coinbase 交易
func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}
	txin := TXInput{[]byte{}, -1, nil, []byte(data)}
	txout := NewTXOutput(subsidy, to)
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}}
	tx.ID = tx.Hash()
	return &tx
}

// 判断是否为铸币交易
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

// 返回交易的哈希值
func (tx *Transaction) Hash() []byte {
	var hash [32]byte
	txCopy := *tx
	txCopy.ID = []byte{}
	hash = sha256.Sum256(txCopy.Serialize())
	return hash[:]
}

// 获取交易副本
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput
	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{vin.Txid, vin.Vout, nil, nil})
	}
	for _, vout := range tx.Vout {
		outputs = append(outputs, TXOutput{vout.Value, vout.PubKeyHash})
	}
	txCopy := Transaction{tx.ID, inputs, outputs}
	return txCopy
}

// 返回一个序列化的交易
func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer
	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	return encoded.Bytes()
}
