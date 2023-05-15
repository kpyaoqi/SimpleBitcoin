package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

const subsidy = 10

// Transaction 由交易 ID，输入和输出构成
type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

// TXInput 包含 3 部分
// Txid: 一个交易输入引用了之前一笔交易的一个输出, ID表明是之前哪笔交易
// Vout: 一笔交易可能有多个输出，Vout 为输出的索引
// ScriptSig: 提供解锁输出 Vout 的数据
type TXInput struct {
	Txid         []byte
	Vout         int
	ScriptPubKey string
}

// TXOutput 包含两部分
// Value: 有多少币，就是存储在 Value 里面
// ScriptPubKey: 对输出进行锁定
// 在当前实现中，ScriptPubKey 将仅用一个字符串来代替
type TXOutput struct {
	Value        int
	ScriptPubKey string
}

// 创建一笔新的交易
func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput
	//找到所有的未花费输出
	acc, vaildOutputs := bc.FindSpendableOutputs(from, amount)
	if acc < amount {
		log.Panic("ERROR:Not enough funds")
	}
	for txid, outs := range vaildOutputs {
		txID, _ := hex.DecodeString(txid)
		for _, out := range outs {
			input := TXInput{txID, out, from}
			inputs = append(inputs, input)
		}
	}

	outputs = append(outputs, TXOutput{amount, to})
	if acc > amount {
		outputs = append(outputs, TXOutput{acc - amount, from})
	}
	tx := Transaction{nil, inputs, outputs}
	tx.SetID()
	return &tx

}

// 构建 coinbase 交易
func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}
	txin := TXInput{[]byte{}, -1, data}
	txout := TXOutput{subsidy, to}
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
	tx.SetID()
	return &tx
}

// 从 address 中找到至少 amount 的 UTXO
func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspenfTXs := bc.FindUnspentTransactions(address)
	accumulated := 0
Work:
	for _, tx := range unspenfTXs {
		txID := hex.EncodeToString(tx.ID)
		for outIdx, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) && accumulated < amount {
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
func (bc *Blockchain) FindUnspentTransactions(address string) []Transaction {
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
				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}
			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(address) {
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

// 判断是否为铸币交易
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte
	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

// unlockingData可以理解为地址
func (in *TXInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptPubKey == unlockingData
}
func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}
