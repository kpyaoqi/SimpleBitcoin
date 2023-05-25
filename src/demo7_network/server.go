package main

//由于我们仅有一个区块链版本，所以 Version 字段实际并不会存储什么重要信息,BestHeight 存储区块链中节点的高度,AddFrom 存储发送者的地址。
type version struct {
	Version    int
	BestHeight int
	AddrFrom   string
}

type getblocks struct {
	AddrFrom string
}

//比特币使用 inv 来向其他节点展示当前节点有什么块和交易 它没有包含完整的区块链和交易，仅仅是哈希而已
//Type 字段表明了这是块还是交易
type inv struct {
	AddrFrom string
	Type     string
	Items    [][]byte
}

type getdata struct {
	AddrFrom string
	Type     string
	ID       []byte
}

type block struct {
	AddrFrom string
	Block    []byte
}

type tx struct {
	AddFrom     string
	Transaction []byte
}

type addr struct {
	AddrList []string
}
