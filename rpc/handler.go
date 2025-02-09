package rpc

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/gagliardetto/solana-go/rpc"
	"github.com/powershitxyz/SolanaProbe/dego"
	"github.com/powershitxyz/SolanaProbe/parser"
	"github.com/powershitxyz/SolanaProbe/pub"
	"github.com/powershitxyz/SolanaProbe/sys"
)

func processBatch(batch []dego.Notification) {
	var programNotifications []dego.ProgramNotification

	for _, notification := range batch {
		method := notification.GetMethod()
		switch method {
		case "programNotification":
			pon, ok := notification.(dego.ProgramNotification)
			if ok {
				logger.Println("ProgramNotification Type: " + pon.Params.Result.Value.Account.Data.Parsed.Type)
			}
		case "slotsUpdatesNotification":
			if pon, ok := notification.(dego.SlotUpdateNotification); ok {
				if pon.Params.Result.Type == "root" {
					slotQueue.Enqueue(pon.Params.Result.Slot)
					timestamp := pon.Params.Result.Timestamp
					if pon.Params.Result.Slot%300 == 0 {
						sys.Logger.Printf("slot: %d ,p: %d  timestamp: %d ,%d", pon.Params.Result.Slot, pon.Params.Result.Parent, timestamp, time.Now().UnixMilli())

					}
				}
			} else {
				logger.Errorf("transfer slot notification error： %v", notification)
			}
		case "slotNotification": // using
			if pon, ok := notification.(dego.SlotNotification); ok {
				slotQueue.Enqueue(pon.Params.Result.Slot)
			} else {
				logger.Errorf("just slot notification error： %v", notification)
			}

		default:
		}
	}

	logger.Println("programNotifications ... ", programNotifications)
}

func ParsingTransactions(decodedTransactions []*dego.RawTransaction, blockWithNoTxs *rpc.GetBlockResult) (res pub.ParsedData, err error) {
	if len(decodedTransactions) == 0 {
		return res, fmt.Errorf("[%d]empty tx list", &blockWithNoTxs.BlockHeight)
	}
	rangeRound := conf.Chain.RangeRound
	batchSlice := pub.BatchSlice(decodedTransactions, rangeRound)
	for i, currentRound := range batchSlice {
		var wg sync.WaitGroup
		var lockTransfer, lockHold, lockFlow, lockPair, lockUpdatePair, pes sync.Mutex
		for j, rawTr := range currentRound {
			wg.Add(1)
			go func(simpleBlock *rpc.GetBlockResult, tx *dego.RawTransaction) {
				defer func() {
					wg.Done()
					if r := recover(); r != nil {
						var txsigs = "txnil weird"
						if tx != nil && tx.Transaction != nil && len(tx.Transaction.Signatures) > 0 {
							txsigs = tx.Transaction.Signatures[0].String()
						}
						logger.Errorf("Panic occurred while processing block %d, tx %s: %v",
							simpleBlock.BlockHeight, txsigs, r)
					}
				}()
				if tx == nil || tx.Transaction == nil || len(tx.Transaction.Signatures) == 0 {
					logger.Errorf("return for tx nil: %v", tx)
					return
				}
				if conf.Chain.GetTxDelay() > 0 {
					source := rand.NewSource(time.Now().UnixNano())
					random := rand.New(source)
					randomDuration := time.Duration(random.Intn(conf.Chain.GetTxDelay())+1) * time.Millisecond
					time.Sleep(randomDuration)
				}
				index := i*rangeRound + j
				resByDecode, err := parser.ParseTransaction(simpleBlock.BlockTime.Time(), tx, index)

				if err != nil {
					logger.Errorf("Parse Block-Tx: %d-%s error： %v", *simpleBlock.BlockHeight, tx.Transaction.Signatures[0].String(), err)
					return
				}

				if len(resByDecode.Transfers) != 0 {
					lockTransfer.Lock()
					res.Transfers = append(res.Transfers, resByDecode.Transfers...)
					lockTransfer.Unlock()
				}
				if len(resByDecode.HoldsUpdateKey) != 0 {
					lockHold.Lock()
					if len(res.HoldsUpdateKey) <= 0 {
						res.HoldsUpdateKey = make(map[int][]string)
					}
					for i2, strings := range resByDecode.HoldsUpdateKey {
						res.HoldsUpdateKey[i2] = strings
					}

					lockHold.Unlock()
				}
				//swap
				if len(resByDecode.TokenFlow) != 0 {
					lockFlow.Lock()
					res.TokenFlow = append(res.TokenFlow, resByDecode.TokenFlow...)
					lockFlow.Unlock()
				}

				if len(resByDecode.NewPairs) != 0 {
					lockPair.Lock()
					res.NewPairs = append(res.NewPairs, resByDecode.NewPairs...)
					lockPair.Unlock()
				}
				if len(resByDecode.UpdatedPairs) != 0 {
					lockUpdatePair.Lock()
					res.UpdatedPairs = append(res.UpdatedPairs, resByDecode.UpdatedPairs...)
					lockUpdatePair.Unlock()
				}
				if len(resByDecode.PoolEvents) != 0 {
					pes.Lock()
					res.PoolEvents = append(res.PoolEvents, resByDecode.PoolEvents...)
					pes.Unlock()
				}
			}(blockWithNoTxs, rawTr)
		}
		wg.Wait()
	}

	if len(res.Transfers) <= 0 && len(res.HoldsUpdateKey) <= 0 && len(res.TokenFlow) <= 0 && len(res.NewPairs) <= 0 && len(res.UpdatedPairs) <= 0 && len(res.PoolEvents) <= 0 {
		logger.Printf("!!!!!!!passok!!!!!!!!!!" + fmt.Sprintf("%d ,[%d] END Transfers :%d , TokenHold :%d, swap :%d, NewPairs :%d, UpPairs :%d ,pes:%d", decodedTransactions[0].Slot, (time.Now().Unix()-blockWithNoTxs.BlockTime.Time().Unix()), len(res.Transfers), len(res.HoldsUpdateKey), len(res.TokenFlow), len(res.NewPairs), len(res.UpdatedPairs), len(res.PoolEvents)))
	}
	return res, nil
}
