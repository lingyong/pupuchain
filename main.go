package main

import (
	"pupuchain/chain"
)

func main() {
	blockchain := chain.NewBlockchain()
	defer blockchain.Db.Close()

	cli := chain.CLI{blockchain}
	cli.Run()
}
