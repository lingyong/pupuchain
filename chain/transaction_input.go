package chain

type TXInput struct {
	Txid []byte
	Vout int
	ScriptSig string
}

