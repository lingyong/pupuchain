package main

import "github.com/boltdb/bolt"

type BlockchainIterator struct {
	currentHash []byte
	db *bolt.DB
}

func (i *BlockchainIterator) Next() *Block {
	var block *Block

	i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		block = DeserialieBlock(encodedBlock)

		return nil
	})

	i.currentHash = block.PrevBlockHash

	return block
}