package blockchain

import "github.com/boltdb/bolt"

const dbFile = "blockchain.db"
const blocksBucket = "blocks"

type BlockChain struct {
	tip []byte
	db *bolt.DB
}

type BlockchainIterator struct {
	 currentHash []byte
	 db 		 *bolt.DB
}

func (chain *BlockChain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{chain.tip, chain.db}

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

func (chain *BlockChain) AddBlock(data string) {
	//prevBlock := chain.Blocks[len(chain.Blocks)-1]
	//newBlock := CreateBlock(data, prevBlock.Hash)
	//chain.Blocks = append(chain.Blocks, newBlock)
	var lastHash []byte

	err := chain.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		panic(err.Error())
	}

	newBlock := CreateBlock(data, lastHash)

	err = chain.db.Update(func(tx *bolt.Tx) error {
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


func InitBlockChain() *BlockChain {
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		panic(err.Error())
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		if b == nil {
			genesis := Genesis()
			b, err = tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				panic(err.Error())
			}
			err = b.Put(genesis.Hash, genesis.Serialize())
			if err != nil {
				panic(err.Error())
			}
			err = b.Put([]byte("l"), genesis.Hash)
			if err != nil {
				panic(err.Error())
			}
			tip = genesis.Hash
		}else {
			tip = b.Get([]byte("l"))
		}
		return nil
	})
	bc := BlockChain{tip, db}
	return &bc
}