package main

import (
	"fmt"
	"crypto/sha256"
	"bytes"
	"encoding/gob"
	"log"
	"encoding/hex"
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

	tx.SetID()

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

func (tx *Transaction) SetID() {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())

	tx.ID = hash[:]
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

func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	acc, validOutputs := bc.FindSpendableOutputs(from, amount)

	if acc < amount {
		log.Panic("Error: not enough funds")
	}

	// for each of the output, build a input for it
	for txid, outs := range validOutputs {
		txId, err := hex.DecodeString(txid)

		if err != nil {
			log.Panic(err)
		}
		for _, out := range outs {
			input := TXInput{txId, out, from}
			inputs = append(inputs, input)
		}
	}

	// build potential two outputs, one will be locked with receiver address, one will be locked with
	// sender address, this is the change.
	outputs = append(outputs, TXOutput{amount, to})
	if acc > amount {
		outputs = append(outputs, TXOutput{acc - amount, from})
	}

	tx := Transaction{nil, inputs, outputs}
	tx.SetID()

	return &tx
}