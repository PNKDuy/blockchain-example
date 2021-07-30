package blockchain

import (
	"encoding/hex"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"os"
)

const dbFile = "blockchain.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

type BlockChain struct {
	tip []byte
	Db  *bolt.DB
}

type BlockchainIterator struct {
	 currentHash []byte
	 db 		 *bolt.DB
}

func (chain *BlockChain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{chain.tip, chain.Db}

	return bci
}

func (i *BlockchainIterator) Next() *Block {
	var block *Block

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		block = DeserializeBlock(encodedBlock)
		return nil
	})
	if err != nil {
		panic(err.Error())
	}

	i.currentHash = block.PrevHash

	return block
}

func (chain *BlockChain) AddBlock(transactions []*Transaction) {
	var lastHash []byte

	err := chain.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		panic(err.Error())
	}

	newBlock := CreateBlock(transactions, lastHash)

	err = chain.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err = b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			panic(err.Error())
		}
		err = b.Put([]byte("l"), newBlock.Hash)
		chain.tip = newBlock.Hash

		return nil
	})

}

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

// NewBlockChain create blockchain with new genesis block
func NewBlockChain(address string) *BlockChain {
	if dbExists() == false {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		panic(err.Error())
	}

	err = db.Update(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	})
	bc := BlockChain{tip, db}
	return &bc
}

// CreateBlockchain creates a new blockchain DB
func CreateBlockchain(address string) *BlockChain {

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
		genesis := Genesis(cbtx)

		b,err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}

		err = b.Put(genesis.Hash, genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), genesis.Hash)
		if err != nil {
			log.Panic(err)
		}

		tip = genesis.Hash
		return nil

	})

	if err != nil {
		log.Panic(err)
	}

	bc := BlockChain{tip: tip, Db: db}
	return &bc
}

func (bc *BlockChain) FindUnspentTransaction(address string) []Transaction{
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transaction {
			txID := hex.EncodeToString(tx.Id)

			Outputs:
				for outIdx, out := range tx.Vout {
					//was the output spent?
					if spentTXOs[txID] != nil {
						for _, spentOut := range spentTXOs[txID] {
							if spentOut == outIdx {
								continue Outputs
							}
						}
					}
					if out.CanBeLockedWith(address) {
						unspentTXs = append(unspentTXs,*tx)
					}
				}
				if tx.IsCoinbase() == false {
					for _, in := range tx.Vin {
						if in.CanUnlockOutputWith(address) {
							inTxID := hex.EncodeToString(in.TxId)
							spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
						}
					}
				}
		}
		if len(block.PrevHash) == 0 {
			break
		}
	}
	return unspentTXs
}

func (bc *BlockChain) FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput
	unspentTransactions := bc.FindUnspentTransaction(address)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.CanBeLockedWith(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}
	return UTXOs
}

func (bc *BlockChain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransaction(address)
	accumulated := 0
	
	Work:
		for _, tx := range unspentTXs {
			txID := hex.EncodeToString(tx.Id)

			for outIdx, out := range tx.Vout {
				if out.CanBeLockedWith(address) && accumulated < amount {
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

