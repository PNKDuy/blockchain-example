package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"time"
)


type Block struct {
	Timestamp int64
	Hash []byte
	Transaction []*Transaction
	PrevHash []byte
	Nonce int
}

//func (b *Block) DeriveHash() {
//	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
//	headers := bytes.Join([][]byte{timestamp, b.Transaction, b.PrevHash}, []byte{})
//	hash := sha256.Sum256(headers)
//	b.Hash = hash[:]
//}

func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.Transaction{
		txHashes = append(txHashes, tx.Id)
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}

func CreateBlock(transactions []*Transaction, prevHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte{}, transactions, prevHash, 0}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}



func Genesis(coinbase *Transaction) *Block {
	return CreateBlock([]*Transaction{coinbase}, []byte{})
}


func (b *Block) Serialize() []byte{
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		panic(err.Error())
	}

	return result.Bytes()
}

func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		panic(err.Error())
	}

	return &block
}
