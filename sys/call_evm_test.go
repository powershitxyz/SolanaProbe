package sys

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"golang.org/x/crypto/sha3"
)

func TestChain(t *testing.T) {

	contract := "0x9333C74BDd1E118634fE5664ACA7a9710b108Bab"
	startBlockInt := uint64(39667601)

	client, err := ethclient.Dial("https://go.getblock.io/e4cb17a6b00a4c3297c5ebdb22a3658d")
	if err != nil {
		log.Println("eventGeneral rpc init:", err)
		return
	}
	finalBlock, err := client.BlockNumber(context.Background())
	if err != nil {
		log.Println("eventGeneral last block:", err)
		return
	}
	log.Println(finalBlock)

	startBlock := new(big.Int).SetUint64(startBlockInt) //deploy block

	lastBlock := new(big.Int).SetUint64(39667601)
	// lastBlock := new(big.Int).SetUint64(finalBlock)
	contractAddress := common.HexToAddress(contract)

	eventSignature := "Transfer(address,address,uint256)"
	hash := sha3.NewLegacyKeccak256()
	hash.Write([]byte(eventSignature))
	eventHash := hash.Sum(nil)

	fmt.Printf("Event signature hash: %s\n", eventHash)
	query := ethereum.FilterQuery{
		FromBlock: startBlock,
		ToBlock:   lastBlock,
		Addresses: []common.Address{
			contractAddress,
		},
		Topics: [][]common.Hash{},
	}
	log.Println("start new event batch:", startBlock.Uint64())
	logs, err := client.FilterLogs(context.Background(), query)
	log.Println("get new event batch log:", startBlock.Uint64(), len(logs))
	if err != nil {
		log.Println("eventGeneral filter:", err)
		return
	}

	log.Println(logs)

}
