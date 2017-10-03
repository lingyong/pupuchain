package main

import (
	"github.com/lingyong/pupuchain/chain"
)

func main() {
	blockchain := chain.NewBlockchain()
	defer blockchain.Db.Close()

	cli := chain.CLI{blockchain}
	cli.Run()
}
