package main

import (
	"github.com/boltdb/bolt"
	"log"
	"os"
	"fmt"
	"encoding/hex"
)

const dbFile = "blockchain.Db"
const blocksBucket = "blocks"

type Blockchain struct {
	tip []byte
	Db  *bolt.DB
}

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

// NewBlockchain creates a new Blockchain with genesis Block
func NewBlockchain() *Blockchain {
	if dbExists() == false {
		fmt.Println("No existing blockchian found. Create one first")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return &Blockchain{tip, db}
}

const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

func CreateBlockchain(address string) *Blockchain {
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		cbtx := NewCoinbaseTx(address, genesisCoinbaseData)
		genesis := NewGenesisBlock(cbtx)

		b, err := tx.CreateBucket([]byte(blocksBucket))
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

	return &Blockchain{tip, db}
}

func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.Db}

	return bci
}

// returns all the transactions that have unspent outputs belong to the address
func (bc *Blockchain) FindUnspentTransactions(address string) []Transaction {
	var unspentTXs []Transaction
	// key is the transaction id, value is the index of the output of the transactions that already spent by the address
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		//  the iterator is going from the most recent block to the genesis block
		block := bci.Next()

		for _, tx := range block.Transactions {
			txId := hex.EncodeToString(tx.ID)

			Outputs:
				for outIndex, out := range tx.Vout {
					// one of the outputs of the transaction is already spent
					if spentTXOs[txId] != nil {
						for _, spentOutputIndex := range spentTXOs[txId] {
							// this means this output in this transaction is already spent
							if spentOutputIndex == outIndex {
								// if the output is already spent, we skip the adding part
								continue Outputs
							}
						}
					}

					// the output is not spent yet and it is belong to the user
					if out.CanUnlockWith(address) {
						unspentTXs = append(unspentTXs, *tx)
					}
				}

				// go over all the transactions' inputs
				if tx.IsCoinbase() == false {
					for _, in := range tx.Vin {
						if in.CanUnlockOutputWith(address) {
							inTxId := hex.EncodeToString(in.Txid)
							spentTXOs[inTxId] = append(spentTXOs[inTxId], in.Vout)
						}
					}
				}

				if len(block.PrevBlockHash) == 0 {
					break
				}
		}

		return unspentTXs
	}

}

func (bc *Blockchain) FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput
	unspentTransactions := bc.FindUnspentTransactions(address)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.CanUnlockWith(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}