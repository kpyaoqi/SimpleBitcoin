package main

import (
	"fmt"
	"log"
	"strconv"
)

// 获取账户余额
func (cli *CLI) getBalance(address string) {
	bc := PositioningBlockchain()
	defer bc.db.Close()
	balance := 0
	//对address解码
	pubKeyHash := Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := bc.FindUTXO(pubKeyHash)
	for _, out := range UTXOs {
		balance += out.Value
	}
	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

// 创建区块链
func (cli *CLI) createBlockchain(address string) {
	if !ValidateAddress(address) {
		log.Panic("ERROR:Address is not valid")
	}
	bc := CreateBlockchain(address)
	bc.db.Close()
	fmt.Println("Done!")
}

// 创建一个钱包
func (cli *CLI) createWallet() {
	wallets, _ := NewWallets()
	address := wallets.CreateWallet()
	wallets.SaveToFile()
	fmt.Printf("Your new address: %s\n", address)
}

// 打印所有钱包地址
func (cli *CLI) listAddresses() {
	wallets, err := NewWallets()
	if err != nil {
		log.Panic(err)
	}
	addresses := wallets.GetAddresses()
	for _, address := range addresses {
		fmt.Println(address)
	}
}

// 发送交易
func (cli *CLI) send(from, to string, amount int) {
	if !ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}
	bc := PositioningBlockchain()
	defer bc.db.Close()
	tx := NewUTXOTransaction(from, to, amount, bc)
	bc.MineBlock([]*Transaction{tx})
	fmt.Println("Success!")
}

// 打印区块链
func (cli *CLI) printChain() {
	bc := PositioningBlockchain()
	defer bc.db.Close()
	bci := bc.Iterator()
	for {
		block := bci.Next()
		fmt.Printf("============ Block %x ============\n", block.Hash)
		fmt.Printf("Prev.block: %x\n", block.PrevBlockHash)
		pow := NewProofOfWork(block)
		fmt.Printf("POW :%s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println("Transactions:")
		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
		fmt.Printf("\n")
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}
