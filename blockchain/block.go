package blockchain

import (
	"bytes"
	"crypto/sha256"
	"strconv"
	"time"
)

type BlockChain struct {
	Blocks []*Block
}

type Block struct {
	Timestamp int64
	Hash []byte
	Data []byte
	PrevHash []byte
	Nonce int
}

func (b *Block) DeriveHash() {
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
	headers := bytes.Join([][]byte{timestamp, b.Data, b.PrevHash}, []byte{})
	hash := sha256.Sum256(headers)
	b.Hash = hash[:]
}

func CreateBlock(data string, prevHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte{}, []byte(data), prevHash, 0}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

func (chain *BlockChain) AddBlock(data string) {
	prevBlock := chain.Blocks[len(chain.Blocks)-1]
	newBlock := CreateBlock(data, prevBlock.Hash)
	chain.Blocks = append(chain.Blocks, newBlock)
}

func Genesis() *Block {
	return CreateBlock("Genesis", []byte{})
}

func InitBlockChain() *BlockChain {
	return &BlockChain{[]*Block{Genesis()}}
}
