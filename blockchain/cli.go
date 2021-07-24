package blockchain

import (
	"flag"
	"os"
)

type CLI struct {
	bc *BlockChain
}

func (cli *CLI) Run() {
	cli.validateArgs()

	addBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	addBlockData := addBlockCmd.String("data", "", "Block data")
	switch os.Args[1] {

	case "addblock":
		{
		err := addBlockCmd.Parse(os.Args[2:])
		if err != nil {
			panic(err.Error())
		}
	}

	case "printchain":
		{
			err := printChainCmd.Parse(os.Args[2:])
			if err != nil {
				panic(err.Error())
			}
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if addBlockCmd.Parsed() {
		if *addBlockData == "" {
			addBlockCmd.Usage()
			os.Exit(1)
		}
		cli.addBlock(*addBlockData)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}
}

func (cli *CLI) addBlock(data string) {
	cli.bc
}