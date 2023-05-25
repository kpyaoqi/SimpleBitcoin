package main

import "crypto/sha256"

type MerkleTree struct {
	RootNode *MerkleNode
}

type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Data  []byte
}

//创建一个新的结点
func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	mNode := MerkleNode{}
	if left == nil && right == nil {
		hash := sha256.Sum256(data)
		mNode.Data = hash[:]
	} else {
		prevHashes := append(left.Data, right.Data...)
		hash := sha256.Sum256(prevHashes)
		mNode.Data = hash[:]
	}
	mNode.Left = left
	mNode.Right = right
	return &mNode
}

//创建一一棵Merkle树
/*func NewMerkleTree(data [][]byte) *MerkleTree {
	var nodes []MerkleNode
	for _, datum := range data {
		node := NewMerkleNode(nil, nil, datum)
		nodes = append(nodes, *node)
	}
	for len(nodes) != 1 {
		nodes = NewMerkleTreeLevel(nodes)
	}
	mTree := MerkleTree{&nodes[0]}
	return &mTree
}*/
func NewMerkleTree(data [][]byte) *MerkleTree {
	var nodes []MerkleNode
	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}
	for _, datum := range data {
		node := NewMerkleNode(nil, nil, datum)
		nodes = append(nodes, *node)
	}
	for i := 0; i < len(data)/2; i++ {
		var newLevel []MerkleNode
		for j := 0; j < len(nodes); j += 2 {
			node := NewMerkleNode(&nodes[j], &nodes[j+1], nil)
			newLevel = append(newLevel, *node)
		}
		nodes = newLevel
	}
	mTree := MerkleTree{&nodes[0]}
	return &mTree
}

//构建MerkleTree的每一层结点
func NewMerkleTreeLevel(nodes []MerkleNode) []MerkleNode {
	if len(nodes)%2 != 0 {
		nodes = append(nodes, nodes[len(nodes)-1])
	}
	var newLevel []MerkleNode
	for j := 0; j < len(nodes); j += 2 {
		node := NewMerkleNode(&nodes[j], &nodes[j+1], nil)
		newLevel = append(newLevel, *node)
	}
	return newLevel
}
