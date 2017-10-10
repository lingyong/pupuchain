package chain

import (
	"fmt"
	"crypto/sha256"
	"bytes"
	"encoding/gob"
	"log"
)

type Transaction struct {
	ID   []byte // id is the hash of the transaction
	Vin  []TXInput
	Vout []TXOutput
}

const subsidy = 10

type TXInput struct {
	Txid      []byte // transaction id of the previous output
	Vout      int // index of the output of the previous transaction
	ScriptSig string
}

type TXOutput struct {
	Value        int
	ScriptPubKey string
}

func NewCoinbaseTx(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to %s", to)
	}

	txin := TXInput{
		Txid:      []byte{},
		Vout:      -1,
		ScriptSig: data,
	}

	txout := TXOutput{
		Value:        subsidy, // amount of reward
		ScriptPubKey: to,
	}

	tx := Transaction{
		ID:   nil,
		Vin:  []TXInput{txin},
		Vout: []TXOutput{txout},
	}

	tx.ID = tx.Hash()

	return &tx
}

func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

// output here means the output of the previous transaction
// this means the address already spent the previous output on this input
func (in *TXInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}

// whether the money store in the output belongs to the user or not
func (out *TXOutput) CanUnlockWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}

// IsCoinbase checks whether the transaction is coinbase
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}