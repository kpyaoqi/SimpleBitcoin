package main

import (
	"fmt"
	"log"
	"strconv"
)

// 获取账户余额
func (cli *CLI) getBalance(address, nodeID string) {
	if !ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := PositioningBlockchain(nodeID)
	UTXOSet := UTXOSet{bc}
	defer bc.db.Close()
	balance := 0
	//对address解码
	pubKeyHash := Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUTXO(pubKeyHash)
	for _, out := range UTXOs {
		balance += out.Value
	}
	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

// 创建区块链
func (cli *CLI) createBlockchain(address, nodeID string) {
	if !ValidateAddress(address) {
		log.Panic("ERROR:Address is not valid")
	}
	bc := CreateBlockchain(address, nodeID)
	defer bc.db.Close()
	UTXOSet := UTXOSet{bc}
	UTXOSet.Reindex()
	fmt.Println("Done!")
}

// 创建一个钱包
func (cli *CLI) createWallet(nodeID string) {
	wallets, _ := NewWallets(nodeID)
	address := wallets.CreateWallet()
	wallets.SaveToFile(nodeID)
	fmt.Printf("Your new address: %s\n", address)
}

// 打印所有钱包地址
func (cli *CLI) listAddresses(nodeID string) {
	wallets, err := NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	addresses := wallets.GetAddresses()
	for _, address := range addresses {
		fmt.Println(address)
	}
}

// 发送交易
func (cli *CLI) send(from, to string, amount int, nodeID string, mineNow bool) {
	if !ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}
	bc := PositioningBlockchain(nodeID)
	UTXOSet := UTXOSet{bc}
	defer bc.db.Close()
	wallets, err := NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)
	tx := NewUTXOTransaction(&wallet, to, amount, &UTXOSet)
	if mineNow {
		cbTX := NewCoinbaseTX(from, "")
		txs := []*Transaction{cbTX, tx}

		newBlock := bc.MineBlock(txs)
		UTXOSet.Update(newBlock)
	} else {
		sendTx(knownNodes[0], tx)
	}

	fmt.Println("Success!")
}

// 打印区块链
func (cli *CLI) printChain(nodeID string) {
	bc := PositioningBlockchain(nodeID)
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
func (cli *CLI) reindexUTXO(nodeID string) {
	bc := PositioningBlockchain(nodeID)
	UTXOSet := UTXOSet{bc}
	UTXOSet.Reindex()

	count := UTXOSet.CountTransactions()
	fmt.Printf("Done! There are %d transactions in the UTXO set.\n", count)
}
func (cli *CLI) startNode(nodeID, minerAddress string) {
	fmt.Printf("Starting node %s\n", nodeID)
	if len(minerAddress) > 0 {
		if ValidateAddress(minerAddress) {
			fmt.Println("Mining is on. Address to receive rewards: ", minerAddress)
		} else {
			log.Panic("Wrong miner address!")
		}
	}
	StartServer(nodeID, minerAddress)
}
