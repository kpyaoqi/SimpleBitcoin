package main

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
)

const protocal = "tcp"

//节点版本
const nodeVersion = 1

//前12字节指定了命令名
const commandLength = 12

var nodeAddress string

//挖矿奖励地址
var miningAddress string

//中心节点
var knownNodes = []string{"localhost:3000"}

//跟踪已下载的块
var blocksInTransit = [][]byte{}

//内存池
var mempool = make(map[string]Transaction)

//开启一个服务器
func StartServer(nodeId, minerAddress string) {
	nodeAddress := fmt.Sprintf("localhost:%s", nodeId)
	miningAddress = minerAddress
	ln, err := net.Listen(protocal, nodeAddress)
	bc := PositioningBlockchain(nodeId)
	if err != nil {
		log.Panicln(err)
	}
	defer ln.Close()
	//这意味着如果当前节点不是中心节点，它必须向中心节点发送 version 消息来查询是否自己的区块链已过时
	if nodeAddress != knownNodes[0] {
		sendVersion(knownNodes[0], bc)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}
		go handleConnection(conn, bc)

	}
}

func gobEncode(data interface{}) []byte {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	enc.Encode(data)
	return buff.Bytes()

}

func commandToBytes(command string) []byte {
	var bytes [commandLength]byte
	for i, c := range command {
		bytes[i] = byte(c)
	}
	return bytes[:]
}

func bytesToCommand(bytes []byte) string {
	var command []byte
	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}
	return fmt.Sprintf("%s", command)
}
func extractCommand(request []byte) []byte {
	return request[:commandLength]
}
func requestBlocks() {
	for _, node := range knownNodes {
		sendGetBlocks(node)
	}
}
func nodeIsKnown(addr string) bool {
	for _, node := range knownNodes {
		if node == addr {
			return true
		}
	}
	return false
}

//发送消息
func sendData(addr string, data []byte) {
	conn, err := net.Dial(protocal, addr)
	if err != nil {
		fmt.Printf("%s is not available\n", addr)
		var updateNodes []string
		for _, node := range knownNodes {
			if node != addr {
				updateNodes = append(updateNodes, node)
			}
		}
		knownNodes = updateNodes
		return
	}
	defer conn.Close()
	io.Copy(conn, bytes.NewReader(data))
}

//当一个节点接收到一个命令，它会运行 bytesToCommand 来提取命令名，并选择正确的处理器处理命令主体
func handleConnection(conn net.Conn, bc *Blockchain) {
	request, _ := ioutil.ReadAll(conn)
	command := bytesToCommand(request[:commandLength])
	fmt.Printf("received %s command\n", command)
	switch command {
	case "addr":
		handleAddr(request)
	case "block":
		handleBlock(request, bc)
	case "inv":
		handleInv(request, bc)
	case "getblocks":
		handleGetBlocks(request, bc)
	case "getdata":
		handleGetData(request, bc)
	case "tx":
		handleTx(request, bc)
	case "version":
		handleVersion(request, bc)
	default:
		fmt.Println("Unknow command!")
	}
	conn.Close()
}

//发送 version 消息
func sendVersion(addr string, bc *Blockchain) {
	bestHeight := bc.GetBestHeight()
	payload := gobEncode(version{nodeVersion, bestHeight, nodeAddress})
	//我们的消息，在底层就是字节序列,前 12 个字节指定了命令名，后面的字节会包含 gob 编码的消息结构
	request := append(commandToBytes("version"), payload...)
	sendData(addr, request)
}
func sendGetBlocks(address string) {
	payload := gobEncode(getblocks{address})
	request := append(commandToBytes("getblocks"), payload...)
	sendData(address, request)
}
func sendInv(address, kind string, blocks [][]byte) {
	payload := gobEncode(inv{address, kind, blocks})
	request := append(commandToBytes("inv"), payload...)
	sendData(address, request)
}
func sendGetData(address, kind string, id []byte) {
	payload := gobEncode(getdata{address, kind, id})
	request := append(commandToBytes("getdata"), payload...)
	sendData(address, request)
}
func sendBlock(address string, b *Block) {
	data := block{nodeAddress, b.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("block"), payload...)
	sendData(address, request)
}
func sendTx(address string, tnx *Transaction) {
	data := tx{nodeAddress, tnx.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("tx"), payload...)
	sendData(address, request)
}
func sendAddr(address string) {
	nodes := addr{knownNodes}
	nodes.AddrList = append(nodes.AddrList, nodeAddress)
	payload := gobEncode(nodes)
	request := append(commandToBytes("addr"), payload...)
	sendData(address, request)
}

//version命令处理器
func handleVersion(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload version
	//对请求进行解码
	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	dec.Decode(&payload)

	myBestHeight := bc.GetBestHeight()
	foreignerBestHeight := payload.BestHeight
	//自身节点的区块链更短
	if myBestHeight < foreignerBestHeight {
		//发送 getblocks 消息
		sendGetBlocks(payload.AddrFrom)
	} else if myBestHeight > foreignerBestHeight {
		//回复 version 消息
		sendVersion(payload.AddrFrom, bc)
	}

	if !nodeIsKnown(payload.AddrFrom) {
		knownNodes = append(knownNodes, payload.AddrFrom)
	}

}

func handleGetBlocks(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload getblocks
	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	dec.Decode(&payload)
	blocks := bc.GetBlockHashes()
	sendInv(payload.AddrFrom, "block", blocks)
}

func handleInv(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload inv
	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	dec.Decode(&payload)
	fmt.Printf("Recevied inventory with %d %s\n", len(payload.Items), payload.Type)
	//在我们的实现中，我们永远也不会发送有多重哈希的 inv
	//这就是为什么当 payload.Type == "tx" 时，只会拿到第一个哈希
	//然后我们检查是否在内存池中已经有了这个哈希，如果没有，发送 getdata 消息。
	if payload.Type == "block" {
		blocksInTransit = payload.Items
		blockHash := payload.Items[0]
		sendGetData(payload.AddrFrom, "block", blockHash)
		newInTransit := [][]byte{}
		for _, b := range blocksInTransit {
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}
		blocksInTransit = newInTransit
	}

	if payload.Type == "tx" {
		txID := payload.Items[0]
		if mempool[hex.EncodeToString(txID)].ID == nil {
			sendGetData(payload.AddrFrom, "tx", txID)
		}
	}
}

func handleGetData(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload getdata
	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	dec.Decode(&payload)

	if payload.Type == "block" {
		block := bc.GetBlock([]byte(payload.ID))
		sendBlock(payload.AddrFrom, &block)
	}

	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := mempool[txID]
		sendTx(payload.AddrFrom, &tx)
	}
}

func handleBlock(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload block

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	dec.Decode(&payload)
	blockData := payload.Block
	block := DeserializeBlock(blockData)

	fmt.Println("Recevied a new block!")
	bc.AddBlock(block)
	fmt.Printf("Added block %x\n", block.Hash)
	//当接收到一个新块时，我们把它放到区块链里面。如果还有更多的区块需要下载，我们继续从上一个下载的块的那个节点继续请求。当最后把所有块都下载完后，对 UTXO 集进行重新索引。
	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		sendGetData(payload.AddrFrom, "block", blockHash)
		blocksInTransit = blocksInTransit[1:]
	} else {
		UTXOSet := UTXOSet{bc}
		UTXOSet.Reindex()
	}
}

func handleTx(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload tx
	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	dec.Decode(&payload)
	txData := payload.Transaction
	tx := DeserializeTransaction(txData)
	//将新交易放到内存池中
	mempool[hex.EncodeToString(tx.ID)] = tx
	if nodeAddress == knownNodes[0] {
		for _, node := range knownNodes {
			if node != nodeAddress && node != payload.AddFrom {
				sendInv(node, "tx", [][]byte{tx.ID})
			}
		}
	} else {
		//如果当前节点（矿工）的内存池中有两笔或更多的交易，开始挖矿
		if len(mempool) >= 2 && len(miningAddress) > 0 {
		MineTransactions:
			var txs []*Transaction
			for id := range mempool {
				tx := mempool[id]
				if bc.VerifyTransaction(&tx) {
					txs = append(txs, &tx)
				}
			}
			if len(txs) == 0 {
				fmt.Println("All transactions are invalid! Waiting for new ones...")
				return
			}
			//如果没有有效交易，则挖矿中断
			cbTx := NewCoinbaseTX(miningAddress, "")
			txs = append(txs, cbTx)
			newBlock := bc.MineBlock(txs)
			UTXOSet := UTXOSet{bc}
			//TODO: 提醒，应该使用 UTXOSet.Update 而不是 UTXOSet.Reindex.
			UTXOSet.Reindex()
			fmt.Println("New block is mined!")
			//当一笔交易被挖出来以后，就会被从内存池中移除。当前节点所连接到的所有其他节点，接收带有新块哈希的 inv 消息。在处理完消息后，它们可以对块进行请求。
			for _, tx := range txs {
				txID := hex.EncodeToString(tx.ID)
				delete(mempool, txID)
			}
			for _, node := range knownNodes {
				if node != nodeAddress {
					sendInv(node, "block", [][]byte{newBlock.Hash})
				}
			}
			if len(mempool) > 0 {
				goto MineTransactions
			}
		}
	}
}

func handleAddr(request []byte) {
	var buff bytes.Buffer
	var payload addr

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	knownNodes = append(knownNodes, payload.AddrList...)
	fmt.Printf("There are %d known nodes now!\n", len(knownNodes))
	requestBlocks()
}
