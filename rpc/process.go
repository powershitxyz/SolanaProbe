package rpc

import (
	"sync"

	"github.com/powershitxyz/SolanaProbe/dego"
	"github.com/powershitxyz/SolanaProbe/model"
	"github.com/powershitxyz/SolanaProbe/pub"
	"github.com/powershitxyz/SolanaProbe/store"
)

func ProcessSlot(slot uint64, lock *sync.WaitGroup) {
	transactions, blockWithNoTxs, err := dego.GetTransactionsBySlot(slot)
	if err != nil {
		logger.Errorf("get slot %d with transactions error： %v", slot, err)
		return
	}
	parsedData, err := ParsingTransactions(transactions, blockWithNoTxs)
	if err != nil {
		logger.Errorf("parse slot %d transactions error： %v", slot, err)
		return
	}

	var wg sync.WaitGroup
	var transferRecordStat, tokenHoldStat, tokenFlowStat, pairStat, updatePairStat int
	if len(parsedData.Transfers) != 0 {
		wg.Add(1)
		go func(ds []model.TransferRecord) {
			defer wg.Done()
			// store.BatchSaveTransferRecord(ds, 500)
			transferRecordStat = len(ds)
		}(parsedData.Transfers)
	}
	//holds
	if len(parsedData.TokenHold) != 0 {
		wg.Add(1)
		go func(ds []model.TokenHold) {
			defer wg.Done()
			//store.BatchSaveOrUpdateHolds(ds, 5000)
			tokenHoldStat = len(ds)
		}(parsedData.TokenHold)
	}
	//swap
	if len(parsedData.TokenFlow) != 0 {
		wg.Add(1)
		go func(ds []model.TokenFlow) {
			defer wg.Done()
			store.BatchSaveTokenFlow(ds, 1000, true)
			tokenFlowStat = len(ds)
		}(parsedData.TokenFlow)
	}
	if len(parsedData.NewPairs) != 0 {
		wg.Add(1)
		go func(ds []pub.Pair) {
			defer wg.Done()
			// store.BatchSavePools(blockWithNoTxs.BlockTime.Time(), ds, 500)
			pairStat = len(ds)
		}(parsedData.NewPairs)
	}
	//更新池子
	if len(parsedData.UpdatedPairs) != 0 {
		wg.Add(1)
		go func(ds []pub.Pair) {
			defer wg.Done()
			// store.BatchUpdatePools(ds, 500)
			updatePairStat = len(ds)
		}(parsedData.UpdatedPairs)
	}
	wg.Wait()
	logger.Printf("----------------Slot: %d, transferRecord: %d, tokenHold: %d, tokenFlow: %d, pair: %d, updatePair: %d",
		slot, transferRecordStat, tokenHoldStat, tokenFlowStat, pairStat, updatePairStat)
	//
}
