package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/powershitxyz/SolanaProbe/database"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/powershitxyz/SolanaProbe/dego"
	"github.com/powershitxyz/SolanaProbe/model"

	"github.com/powershitxyz/SolanaProbe/parser/jup"
	"github.com/powershitxyz/SolanaProbe/parser/other"
	"github.com/powershitxyz/SolanaProbe/parser/raydium"
	"github.com/powershitxyz/SolanaProbe/pub"
	"github.com/powershitxyz/SolanaProbe/sys"
	"github.com/shopspring/decimal"
)

var Log = sys.Logger

var lookupTableCache = make(map[solana.PublicKey]*TempLookupTable)
var lookupMu sync.RWMutex

func ParseTest(txTran *solana.Transaction, txMeta *rpc.TransactionMeta) {
	ParseTransaction(time.Now(), &dego.RawTransaction{
		Slot:        0,
		Transaction: txTran,
		Meta:        txMeta,
	}, 1)

}

type TempLookupTable struct {
	Times      int64
	Time       time.Time
	Addresses  solana.PublicKeySlice
	AccountKey solana.PublicKey
}

func ParseTransaction(txTime time.Time, decodedTx *dego.RawTransaction, index int) (res pub.ParsedData, err error) {
	defer func() {
		if r := recover(); r != nil {
			Log.Errorf("entry.go ParseTransaction Recovered %v", r)
		}
	}()
	txMeta := decodedTx.Meta
	txSlot := decodedTx.Slot
	txTran := decodedTx.Transaction
	txSender := txTran.Message.Signers()[0].String()
	if txSlot > 0 && txSlot%12000 == 0 {
		var toDelete []solana.PublicKey
		lookupMu.Lock()
		for k, v := range lookupTableCache {
			if time.Now().Unix()-v.Time.Unix() > 6000 {
				toDelete = append(toDelete, k)
			}
		}
		for _, k := range toDelete {
			delete(lookupTableCache, k)
		}
		toDelete = nil
		lookupMu.Unlock()
	}

	// 获取 Address Lookup Table 的公共密钥
	lookupTableAccounts := txTran.Message.AddressTableLookups

	table := make(map[solana.PublicKey]solana.PublicKeySlice)
	foundLookupTable := false
	for _, lookup := range lookupTableAccounts {
		// 检查缓存中是否已有该账户的查找表结果
		lookupMu.RLock()
		temped, found := lookupTableCache[lookup.AccountKey]
		foundLookupTable = found
		lookupMu.RUnlock()
		if found && len(temped.Addresses) > 0 {
			table[lookup.AccountKey] = temped.Addresses
			continue
		}

		if !found || temped.Times < 5 {
			lookupTable, err := dego.GetAddressLookupTableWithRetry(context.Background(), lookup.AccountKey)
			if err == nil {
				table[lookup.AccountKey] = lookupTable.Addresses
				lookupMu.Lock()
				times := int64(1)
				if found {
					times = temped.Times + 1
				}
				lookupTableCache[lookup.AccountKey] = &TempLookupTable{
					Times:      times,
					AccountKey: lookup.AccountKey,
					Addresses:  lookupTable.Addresses,
					Time:       time.Now(),
				}
				lookupMu.Unlock()
			} else {
				if !strings.Contains(err.Error(), "not found") {
					Log.Errorf("Lookup Table Error 没有找到地址表 : %s:%s , %v", lookup.AccountKey.String(), txTran.Signatures[0].String(), err)
				}

				lookupMu.Lock()
				times := int64(1)
				if found {
					times = temped.Times + 1
				}
				lookupTableCache[lookup.AccountKey] = &TempLookupTable{
					Times:      times,
					AccountKey: lookup.AccountKey,
					Addresses:  nil,
					Time:       time.Now(),
				}
				lookupMu.Unlock()
				break
			}
		}

	}

	if err != nil {
		return res, err
	}
	if len(table) > 0 {
		// 初始化地址表
		txTran.Message.SetAddressTables(table)
	}
	// 获取完整的账户列表
	metas, err := txTran.Message.AccountMetaList()
	if err != nil {
		Log.Errorf("Get Writable Keys Error  Chain? :%v  %s:%v", !foundLookupTable, txTran.Signatures[0].String(), err)
		//缓存拿的LookupTable 可能有问题 重新从链上获取
		if foundLookupTable && strings.Contains(err.Error(), "address table lookup index out of") {
			for _, lookup := range lookupTableAccounts {
				// 检查缓存中是否已有该账户的查找表结果

				lookupTable, err := dego.GetAddressLookupTableWithRetry(context.Background(), lookup.AccountKey)
				if err == nil {
					//放入table
					table[lookup.AccountKey] = lookupTable.Addresses
					//拿到缓存lock
					lookupMu.Lock()
					times := int64(1)
					//链上获取的地址表 放入本地缓存
					lookupTableCache[lookup.AccountKey] = &TempLookupTable{
						Times:      times,
						AccountKey: lookup.AccountKey,
						Addresses:  lookupTable.Addresses,
						Time:       time.Now(),
					}
					lookupMu.Unlock()
				} else {
					// 搞mev的叼毛把地址表关闭了
					if !strings.Contains(err.Error(), "not found") {
						Log.Errorf("Lookup Table Error 没有找到地址表  : %s:%s , %v", lookup.AccountKey.String(), txTran.Signatures[0].String(), err)
					}

					lookupMu.Lock()
					times := int64(1)
					//链上获取的地址表 放入本地缓存
					lookupTableCache[lookup.AccountKey] = &TempLookupTable{
						Times:      times,
						AccountKey: lookup.AccountKey,
						Addresses:  nil,
						Time:       time.Now(),
					}
					lookupMu.Unlock()
					break
				}

			}
			// 重新设置地址表
			if len(table) > 0 {
				txTran.Message.SetAddressTables(table)
			}
			// 重新获取账户列表 这次失败就是真的失败了直接返回nil
			metas, err = txTran.Message.AccountMetaList()
			if err != nil {
				Log.Errorf("Get Writable Keys Error inChain :  %s:%v", txTran.Signatures[0].String(), err)
				return res, err
			}
		}

	}
	createAccounts := make([]*DecodeInsData, 0)
	liqSli := make([]*DecodeInsData, 0)
	transfers := make([]*token.Instruction, 0)
	sysTransfers := make([]*system.Instruction, 0)
	transferChecked := make([]*token.Instruction, 0)
	//BurnChecked := make([]*token.BurnChecked, 0)
	varInners := txMeta.InnerInstructions

	ths := make([]model.TokenHold, 0)
	holdUpdateKeys := make([]string, 0)
	balanceChg := make([]model.BalanceChange, 0)
	tokenDecimalsMap := make(map[string]uint8)
	// accountMap 临时account
	accountMap := make(map[string]pub.TempAccountData)
	solHolds := make(map[string]uint64)
	//获取token余额
	//收集代币精度 tokenDecimalsMap
	ths, solHolds, holdUpdateKeys, balanceChg = processTempAcountsAndBalance(txSlot, txTime, txMeta, metas, tokenDecimalsMap, ths, accountMap)
	okxMsgPubStatus(txSlot, txTime, txSender, txTran, txMeta, tokenDecimalsMap, balanceChg)

	for i, _ := range txTran.Message.Instructions {
		varr := txTran.Message.Instructions[i]

		programPub, _ := txTran.Message.ResolveProgramIDIndex(varr.ProgramIDIndex)
		accounts := make([]*solana.AccountMeta, 0)
		for _, v := range varr.Accounts {
			accounts = append(accounts, metas[v])
		}
		var insInners *[]solana.CompiledInstruction
		var insLen = 0
		for j, _ := range varInners {
			if varInners[j].Index == uint16(i) {
				insInners = &varInners[j].Instructions
				insLen = len(varInners[j].Instructions)
			}
		}

		result, err := ParseDecode(programPub, accounts, varr.Data, insInners, metas, accountMap, txTran.Signatures[0].String())
		if err != nil && err.Error() == "no decoder found" && insLen >= 2 {
			result, err = ParseOtherDecode(programPub, accounts, varr.Data, insInners, metas, accountMap, txTran.Signatures[0].String())
		}
		if err != nil {
			//Log.Println(err)
		} else {
			switch result.Key {
			case pub.Transfer:
				transfers = append(transfers, result.Data.(*token.Instruction))
			case pub.SysTransfer:
				sysTransfers = append(sysTransfers, result.Data.(*system.Instruction))
			case pub.TransferChecked:
				transferChecked = append(transferChecked, result.Data.(*token.Instruction))
			//case pub.BurnChecked:
			//BurnChecked = append(BurnChecked, result.Data.(*token.BurnChecked))
			case pub.InitializeAccount, pub.InitializeAccount2, pub.InitializeAccount3:
				createAccounts = append(createAccounts, result)
			default:
				//swap
				if strings.HasPrefix(result.Key, "Liq") {
					liqSli = append(liqSli, result)
				}
			}
		}
	}

	processTempAccountMap(txSlot, txTime, tokenDecimalsMap, createAccounts, accountMap)

	//spew.Dump("createAccounts  " + fmt.Sprintf("%d", len(transfers)) + " " + fmt.Sprintf("%d", len(sysTransfers)) + "  ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	trs := processTransferRecordEntry(pub.DecodeInsTempData{
		TxSlot:           txSlot,
		TxTime:           txTime,
		Tx:               txTran.Signatures[0].String(),
		TempAccount:      accountMap,
		TokenDecimalsMap: tokenDecimalsMap,
		Fee:              txMeta.Fee,
	}, transfers, sysTransfers, transferChecked)

	tfs, newPairs, updatePairs, poolEvents := processSwapEntry(pub.DecodeInsTempData{
		TxSlot:           txSlot,
		TxTime:           txTime,
		Tx:               txTran.Signatures[0].String(),
		TempAccount:      accountMap,
		TokenDecimalsMap: tokenDecimalsMap,
		Fee:              txMeta.Fee,
		TokenHolds:       ths,
		SOLHolds:         solHolds,
	}, liqSli)
	// 转账总量 in out
	//for _, tr := range trs {
	//	out := mycache.HoldUpdateInOut(tr)
	//	holdUpdateKeys = append(holdUpdateKeys, out...)
	//}
	// 交易总量
	//for _, tr := range tfs {
	//	out := mycache.HoldUpdateInOutSwap(tr)
	//	holdUpdateKeys = append(holdUpdateKeys, out...)
	//}
	res.Transfers = trs
	res.TokenHold = ths
	if len(holdUpdateKeys) > 0 {
		var tempM = make(map[int][]string)
		tempM[index] = holdUpdateKeys
		res.HoldsUpdateKey = tempM
	}
	balanceChgMap := make(map[string]model.BalanceChange)
	for _, change := range balanceChg {
		t := change.Token
		if "SOL" == change.Token {
			t = pub.WSOL
		}
		balanceChgMap[change.Owner+t] = change
	}
	tfs1 := make([]model.TokenFlow, 0)
	for _, tf := range tfs {

		so := make([]model.BalanceChange, 0)
		change, exist := balanceChgMap[txSender+tf.Token0]
		if exist {
			so = append(so, change)
		}
		change1, exist1 := balanceChgMap[txSender+tf.Token1]
		if exist1 {
			so = append(so, change1)
		}
		if len(so) > 0 {
			tf.Chg = so
		}

		tfs1 = append(tfs1, tf)

	}
	res.TokenFlow = tfs1
	i := len(newPairs)
	if i > 0 {
		pair := newPairs[i-1]
		pair.CreatedAt = txTime.Unix()
		newPairs[i-1] = pair
	}
	res.NewPairs = newPairs
	res.UpdatedPairs = updatePairs
	res.PoolEvents = poolEvents

	//if len(res.Transfers) > 0 || len(res.TokenHold) > 0 {
	//	Log.Printf("Transfer-TokenHold Len:  " + fmt.Sprintf("res.Transfers :%d , res.TokenHold :%d, res.TokenFlow :%d, res.NewPairs :%d, res.UpdatedPairs :%d tx: %s  ", len(res.Transfers), len(res.TokenHold), len(res.TokenFlow), len(res.NewPairs), len(res.UpdatedPairs), txTran.Signatures[0].String()))
	//}
	return res, nil
}

func okxMsgPubStatus(slot uint64, txTime time.Time, sender string, tran *solana.Transaction, meta *rpc.TransactionMeta, tokenDecimalsMap map[string]uint8, balanceChg []model.BalanceChange) {

	msgArr := make([]string, 0)
	okxMsgFlag := 0
	token0 := ""
	token1 := ""
	tokenFlag := 0
	for _, msg := range meta.LogMessages {
		if strings.HasPrefix(msg, "Program log: order_id") {
			msgArr = append(msgArr, msg)
			okxMsgFlag++
			tokenFlag = 1
			continue
		}
		if tokenFlag == 1 {
			split := strings.Split(msg, ":")
			token0 = strings.ReplaceAll(split[1], " ", "")
			tokenFlag = 2
			continue
		}
		if tokenFlag == 2 {
			split := strings.Split(msg, ":")
			token1 = strings.ReplaceAll(split[1], " ", "")
			tokenFlag = 3
			continue
		}
		if strings.HasPrefix(msg, "Program log: before_source_balance:") {
			msgArr = append(msgArr, msg)
			okxMsgFlag++
			continue
		}
		if strings.HasPrefix(msg, "Program log: after_source_balance:") {
			msgArr = append(msgArr, msg)
			okxMsgFlag++
			continue
		}
	}
	if okxMsgFlag == 3 && len(msgArr) == 3 {
		before := msgArr[1]
		after := msgArr[2]
		splitBefore := strings.Split(before, ",")
		splitAfter := strings.Split(after, ",")
		amount_in_before := strings.ReplaceAll(strings.ReplaceAll(splitBefore[2], "amount_in:", ""), " ", "")
		amount_out_before := strings.ReplaceAll(strings.ReplaceAll(splitBefore[3], "expect_amount_out:", ""), " ", "")
		amount_in_after := strings.ReplaceAll(strings.ReplaceAll(splitAfter[2], "source_token_change:", ""), " ", "")
		amount_out_after := strings.ReplaceAll(strings.ReplaceAll(splitAfter[3], "destination_token_change:", ""), " ", "")
		amount_inStr := amount_in_after
		amount_outStr := amount_out_after
		if amount_in_after == "0" || amount_in_after == "" {
			amount_inStr = amount_in_before
		}
		if amount_out_after == "0" || amount_out_after == "" {
			amount_outStr = amount_out_before
		}
		amount_in, _ := decimal.NewFromString(amount_inStr)
		amount_out, _ := decimal.NewFromString(amount_outStr)
		token0Decimals, exist0 := tokenDecimalsMap[token0]
		token1Decimals, exist1 := tokenDecimalsMap[token1]
		if !exist0 || !exist1 {
			pub.Log.Println("拿不到token精度 okxMsgPubStatus tokenDecimalsMap is nil ," + tran.Signatures[0].String())
			return
		}
		quoteToken0, e0 := QuoteMap[token0]
		quoteToken1, e1 := QuoteMap[token1]
		if !e0 && !e1 {
			return
		}
		price := decimal.Zero
		amount_in_decimal := amount_in.Div(decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(token0Decimals))))
		amount_out_decimal := amount_out.Div(decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(token1Decimals))))
		if e0 && e1 {
			//token0是sol 是山寨币
			if token0 == pub.SOL || token0 == pub.SOLStr || token0 == pub.WSOL {
				price = quoteToken0.Price
			}
			if token1 == pub.SOL || token1 == pub.SOLStr || token1 == pub.WSOL {
				price = quoteToken0.Price
			}
			if (token1 == pub.USDT || token1 == pub.USDC) && (token0 == pub.USDT || token0 == pub.USDC) {
				price = decimal.NewFromInt32(1)
			}
		}
		if e0 && !e1 {
			//token0 是报价币  token1是山寨币
			price = quoteToken0.Price.Mul(amount_in_decimal).Div(amount_out_decimal)
		}
		if !e0 && e1 {
			//token1 是报价币  token0是山寨币
			price = quoteToken1.Price.Mul(amount_out_decimal).Div(amount_in_decimal)
		}
		var t0Change model.BalanceChange
		var t1Change model.BalanceChange
		t0Str := token0
		t1Str := token1
		if t0Str == pub.SOL || t0Str == pub.SOLStr || t0Str == pub.WSOL {
			t0Str = pub.SOLStr
		}
		if t1Str == pub.SOL || t1Str == pub.SOLStr || t1Str == pub.WSOL {
			t1Str = pub.SOLStr
		}
		for _, change := range balanceChg {
			if change.Owner == sender && change.Token == t0Str {
				t0Change = change
			}
			if change.Owner == sender && change.Token == t1Str {
				t1Change = change
			}
		}
		if !price.IsZero() {
			swapInfo := model.OkxSwapInfo{
				Tx:                tran.Signatures[0].String(),
				Sender:            sender,
				TokenFrom:         token0,
				TokenTo:           token1,
				AmountFrom:        amount_in,
				AmountTo:          amount_out,
				TokenFromDecimals: token0Decimals,
				TokenToDecimals:   token1Decimals,
				Price:             price,
				Block:             slot,
				BlockTime:         txTime.Unix(),
				PubTime:           time.Now().Unix(),
				PubTimeStr:        time.Now().Format("2006-01-02 15:04:05.000"),
				TokenFromChg:      t0Change,
				TokenToChg:        t1Change,
				GasUsed:           meta.Fee,
			}
			js, _ := json.Marshal(swapInfo)
			//pub.Log.Println("okxMsgPubStatus  " + string(js))
			database.PublishS(pub.OkxTradeTopic, js)
		} else {
			pub.Log.Println("拿不到报价	okxMsgPubStatus price is 0," + tran.Signatures[0].String())
		}
	}

}

func processSwapEntry(tempData pub.DecodeInsTempData, liqSli []*DecodeInsData) (tfs []model.TokenFlow, newPairsRes []pub.Pair, updatePairsRes []pub.Pair, poolsEvent []model.PoolEvent) {
	if len(liqSli) == 0 {
		return tfs, newPairsRes, updatePairsRes, poolsEvent
	}
	holds := tempData.TokenHolds
	MyHolds := make(map[string]decimal.Decimal)
	for _, hold := range holds {
		MyHolds[hold.Account] = hold.Amount
	}
	sps := make([]pub.SwapInfo, 0)
	for _, insData := range liqSli {
		switch insData.Plat {
		case pub.DexRaydium:
			switch insData.Key {
			case pub.LiqSwap:
				swap := insData.Data.(*raydium.RaydiumLiqSwap)
				if swap.InAddr == nil {
					fmt.Println(tempData.Tx, "InAddr is nil")
				}
				spsByjup := SwapRaydium(tempData.TxSlot, tempData.TxTime, swap, tempData.TokenDecimalsMap, pub.DexRaydium, tempData.Tx)
				if len(spsByjup) > 0 {
					sps = append(sps, spsByjup...)
				}

			case pub.LiqAdd:
				poolEvent, uPairs, nPairs := liqAddWithRayDium(tempData.TxTime, tempData.TxSlot, holds, tempData.TokenDecimalsMap, insData.Data, pub.DexRaydium, tempData.Tx, pub.LiqAdd)
				newPairsRes = append(newPairsRes, nPairs...)
				updatePairsRes = append(updatePairsRes, uPairs...)
				poolsEvent = append(poolsEvent, poolEvent...)
			case pub.LiqRemove:
				poolEvent, uPairs, nPairs := liqAddWithRayDium(tempData.TxTime, tempData.TxSlot, holds, tempData.TokenDecimalsMap, insData.Data, pub.DexRaydium, tempData.Tx, pub.LiqRemove)
				newPairsRes = append(newPairsRes, nPairs...)
				updatePairsRes = append(updatePairsRes, uPairs...)
				poolsEvent = append(poolsEvent, poolEvent...)
			case pub.LiqCreate:

				poolEvent, uPairs, nPairs := liqAddWithRayDium(tempData.TxTime, tempData.TxSlot, holds, tempData.TokenDecimalsMap, insData.Data, pub.DexRaydium, tempData.Tx, pub.LiqCreate)
				newPairsRes = append(newPairsRes, nPairs...)
				updatePairsRes = append(updatePairsRes, uPairs...)
				poolsEvent = append(poolsEvent, poolEvent...)
			}
		case pub.DexOrca:
			switch insData.Key {
			case pub.LiqSwap:
				swap := insData.Data.(*raydium.RaydiumLiqSwap)

				spsByjup := SwapRaydium(tempData.TxSlot, tempData.TxTime, swap, tempData.TokenDecimalsMap, pub.DexOrca, tempData.Tx)
				if len(spsByjup) > 0 {
					sps = append(sps, spsByjup...)
				}

			case pub.LiqAdd:
				poolEvent, uPairs, nPairs := liqAddWithRayDium(tempData.TxTime, tempData.TxSlot, holds, tempData.TokenDecimalsMap, insData.Data, pub.DexOrca, tempData.Tx, pub.LiqAdd)
				newPairsRes = append(newPairsRes, nPairs...)
				updatePairsRes = append(updatePairsRes, uPairs...)
				poolsEvent = append(poolsEvent, poolEvent...)
			case pub.LiqRemove:
				poolEvent, uPairs, nPairs := liqAddWithRayDium(tempData.TxTime, tempData.TxSlot, holds, tempData.TokenDecimalsMap, insData.Data, pub.DexOrca, tempData.Tx, pub.LiqRemove)
				newPairsRes = append(newPairsRes, nPairs...)
				updatePairsRes = append(updatePairsRes, uPairs...)
				poolsEvent = append(poolsEvent, poolEvent...)
			case pub.LiqCreate:

				poolEvent, uPairs, nPairs := liqAddWithRayDium(tempData.TxTime, tempData.TxSlot, holds, tempData.TokenDecimalsMap, insData.Data, pub.DexRaydium, tempData.Tx, pub.LiqCreate)
				newPairsRes = append(newPairsRes, nPairs...)
				updatePairsRes = append(updatePairsRes, uPairs...)
				poolsEvent = append(poolsEvent, poolEvent...)
			}
		case pub.DexMeteora:
			switch insData.Key {
			case pub.LiqSwap:
				swap := insData.Data.(*raydium.RaydiumLiqSwap)

				spsByjup := SwapRaydium(tempData.TxSlot, tempData.TxTime, swap, tempData.TokenDecimalsMap, pub.DexOrca, tempData.Tx)
				if len(spsByjup) > 0 {
					sps = append(sps, spsByjup...)
				}

			case pub.LiqAdd:
				poolEvent, uPairs, nPairs := liqAddWithRayDium(tempData.TxTime, tempData.TxSlot, holds, tempData.TokenDecimalsMap, insData.Data, pub.DexOrca, tempData.Tx, pub.LiqAdd)
				newPairsRes = append(newPairsRes, nPairs...)
				updatePairsRes = append(updatePairsRes, uPairs...)
				poolsEvent = append(poolsEvent, poolEvent...)
			case pub.LiqRemove:
				poolEvent, uPairs, nPairs := liqAddWithRayDium(tempData.TxTime, tempData.TxSlot, holds, tempData.TokenDecimalsMap, insData.Data, pub.DexOrca, tempData.Tx, pub.LiqRemove)
				newPairsRes = append(newPairsRes, nPairs...)
				updatePairsRes = append(updatePairsRes, uPairs...)
				poolsEvent = append(poolsEvent, poolEvent...)
			case pub.LiqCreate:

				poolEvent, uPairs, nPairs := liqAddWithRayDium(tempData.TxTime, tempData.TxSlot, holds, tempData.TokenDecimalsMap, insData.Data, pub.DexRaydium, tempData.Tx, pub.LiqCreate)
				newPairsRes = append(newPairsRes, nPairs...)
				updatePairsRes = append(updatePairsRes, uPairs...)
				poolsEvent = append(poolsEvent, poolEvent...)
			}
		case pub.DexJupiterV6:
			jupSw := insData.Data.(*jup.JupSwap)
			data := jupSw.Data

			for _, swap := range data {
				if swap.InAddr == nil {
					fmt.Println(tempData.Tx, "InAddr is nil")
				}
				spsByjup := SwapRaydium(tempData.TxSlot, tempData.TxTime, swap, tempData.TokenDecimalsMap, pub.DexJupiterV6, tempData.Tx)
				if len(spsByjup) > 0 {
					sps = append(sps, spsByjup...)
				}
			}
		case pub.DexOther:
			jupSw := insData.Data.(*other.OtherSwap)
			data := jupSw.Data

			for _, swap := range data {
				if swap.InAddr == nil {
					fmt.Println(tempData.Tx, "InAddr is nil")
				}
				spsByjup := SwapRaydium(tempData.TxSlot, tempData.TxTime, swap, tempData.TokenDecimalsMap, pub.DexOther, tempData.Tx)
				if len(spsByjup) > 0 {
					sps = append(sps, spsByjup...)
				}
			}
		}
		// case pub.DexOkxProxy:
		// 	jupSw := insData.Data.(*okxproxy.OkxProxy)
		// 	data := jupSw.Data

		// 	for _, swap := range data {
		// 		if swap.InAddr == nil {
		// 			fmt.Println(tempData.Tx, "InAddr is nil")
		// 		}
		// 		spsByjup := SwapRaydium(tempData.TxSlot, tempData.TxTime, swap, tempData.TokenDecimalsMap, pub.DexOther, tempData.Tx)
		// 		if len(spsByjup) > 0 {
		// 			sps = append(sps, spsByjup...)
		// 		}
		// 	}
		// }

	}

	if len(sps) > 0 {
		for _, sp := range sps {
			// _, _, updatePairs, newPairs := checkAndGetPair(holds, tempData.SOLHolds, sp)
			// //pairInCache, exist, updatePairs, newPairs := checkAndGetPair(MyHolds, sp)
			// newPairsRes = append(newPairsRes, newPairs...)
			// updatePairsRes = append(updatePairsRes, updatePairs...)

			token0 := sp.BaseToken
			token1 := sp.QuoteToken

			var amount0In, amount0Out, amount1In, amount1Out decimal.Decimal
			var token0Decimals, token1Decimals uint8
			if sp.In == sp.BaseToken {
				amount0In = sp.AmountIn
				token0Decimals = sp.BaseDecimals
				token1Decimals = sp.QuoteDecimals

				amount0Out = decimal.Zero
				amount1In = decimal.Zero
				amount1Out = sp.AmountOut
			} else {
				amount0In = decimal.Zero
				amount0Out = sp.AmountOut
				amount1In = sp.AmountIn
				amount1Out = decimal.Zero

				token1Decimals = sp.BaseDecimals
				token0Decimals = sp.QuoteDecimals
			}

			tfs = append(tfs, model.TokenFlow{
				Time:           tempData.TxTime,
				Pair:           sp.Pair,
				Type:           sp.Dex,
				From:           sp.From,
				Block:          tempData.TxSlot,
				Token0:         token0,
				Token1:         token1,
				Amount0In:      amount0In,
				Amount0Out:     amount0Out,
				Amount1In:      amount1In,
				Amount1Out:     amount1Out,
				GasUse:         tempData.Fee,
				TxTime:         tempData.TxTime.Unix(),
				Price:          sp.Price,
				ChainCode:      "SOLANA",
				Sender:         sp.From,
				Tx:             sp.TxHash,
				Token0Decimals: token0Decimals,
				Token1Decimals: token1Decimals,
			})
		}
	}
	return tfs, newPairsRes, updatePairsRes, poolsEvent
}

func checkAndUpdatePair(txTime time.Time, holds []model.TokenHold, tDecimalsMap map[string]uint8, sp model.PoolEvent, tx string) (pairInCache *pub.Pair, exi bool, updatePairs []pub.Pair, newPairs []pub.Pair) {
	if sp.Dex == pub.DexPumpFun || sp.Dex == pub.DexMoonshot {
		return nil, false, updatePairs, newPairs
	}
	updatePairs = make([]pub.Pair, 0)
	newPairs = make([]pub.Pair, 0)
	MyHolds1 := make(map[string]decimal.Decimal)
	MyHoldsMint := make(map[string]string)
	for _, hold := range holds {
		MyHolds1[hold.Account] = hold.Amount
		MyHoldsMint[hold.Account] = hold.TokenAddress
	}
	coinAccount := sp.CoinAccount
	pcAccount := sp.PcAccount
	addr0, exAddr0 := MyHoldsMint[coinAccount]
	addr1, exAddr1 := MyHoldsMint[pcAccount]
	if !exAddr0 || !exAddr1 {
		return nil, false, updatePairs, newPairs
	}
	baseDecimals := tDecimalsMap[addr0]
	quoteDecimals := tDecimalsMap[addr1]
	quoteToken0, ok0 := QuoteMap[addr0]
	quoteToken1, ok1 := QuoteMap[addr1]
	quoteAddr := addr1
	baseAddr := addr0
	if !ok0 && !ok1 {
		return nil, false, updatePairs, newPairs
	}
	fmt.Println(baseDecimals, quoteDecimals, quoteAddr, baseAddr)
	if ok0 && ok1 {
		if (quoteToken0.PairSymbol == "USDC") != (quoteToken1.PairSymbol == "USDC") {
			//1 为主币 0为价值币
			if quoteToken0.PairSymbol == "USDC" {
				coinAccount = sp.PcAccount
				pcAccount = sp.CoinAccount
				baseDecimals = tDecimalsMap[addr1]
				quoteDecimals = tDecimalsMap[addr0]
				quoteAddr = addr0
				baseAddr = addr1
			}
		}
	} else if ok0 && !ok1 {
		//0为价值币
		coinAccount = sp.PcAccount
		pcAccount = sp.CoinAccount
		baseDecimals = tDecimalsMap[addr1]
		quoteDecimals = tDecimalsMap[addr0]
		quoteAddr = addr0
		baseAddr = addr1
	}

	return pairInCache, false, updatePairs, newPairs
}

func liqAddWithRayDium(txTime time.Time, slot uint64, MyHolds []model.TokenHold, tDecimalsMap map[string]uint8, eventData interface{}, dexRaydium string, tx string, event string) (poolE []model.PoolEvent, updatePairs []pub.Pair, newPairs []pub.Pair) {

	switch event {
	case pub.LiqAdd:
		liq := eventData.(*raydium.RaydiumLiqAdd)
		poolEvent := model.PoolEvent{
			Pair:                liq.LpPair,
			Token0:              liq.CoinAddr,
			Token1:              liq.PcAddr,
			Amount0:             decimal.NewFromBigInt(new(big.Int).SetUint64(liq.MaxCoinAmount), 0),
			Amount1:             decimal.NewFromBigInt(new(big.Int).SetUint64(liq.MaxPcAmount), 0),
			Liquidity:           decimal.NewFromBigInt(&liq.Liq, 0),
			AddLiquidityAddress: liq.Authority,
			Block:               slot,
			Tx:                  tx,
			Dex:                 liq.Dex,
			Type:                pub.LIQ_ADD,
			TxTime:              txTime.Unix(),
			ChainCode:           pub.SOLANA,
			CoinAccount:         liq.CoinAccount,
			PcAccount:           liq.PcAccount,
		}
		_, _, uPairs, nPairs := checkAndUpdatePair(txTime, MyHolds, tDecimalsMap, poolEvent, tx)
		poolE = append(poolE, poolEvent)
		updatePairs = append(updatePairs, uPairs...)
		newPairs = append(newPairs, nPairs...)
	case pub.LiqRemove:
		liq := eventData.(*raydium.RaydiumLiqRemove)

		//
		poolEvent := model.PoolEvent{
			Pair:                liq.LpPair,
			Token0:              liq.CoinAddr,
			Token1:              liq.PcAddr,
			Amount0:             decimal.NewFromBigInt(new(big.Int).SetUint64(liq.CoinAmount), 0),
			Amount1:             decimal.NewFromBigInt(new(big.Int).SetUint64(liq.PcAmount), 0),
			Liquidity:           decimal.NewFromBigInt(new(big.Int).SetUint64(0), 0),
			AddLiquidityAddress: liq.Authority,
			Block:               slot,
			Tx:                  tx,
			Dex:                 liq.Dex,
			Type:                pub.LIQ_REMOVE,
			TxTime:              txTime.Unix(),
			ChainCode:           pub.SOLANA,
			CoinAccount:         liq.CoinAccount,
			PcAccount:           liq.PcAccount,
		}
		_, _, uPairs, nPairs := checkAndUpdatePair(txTime, MyHolds, tDecimalsMap, poolEvent, tx)
		poolE = append(poolE, poolEvent)
		updatePairs = append(updatePairs, uPairs...)
		newPairs = append(newPairs, nPairs...)
	case pub.LiqCreate:
		liq := eventData.(*raydium.RaydiumLiqCreate)
		poolEvent := model.PoolEvent{
			Pair:                liq.LpPair,
			Token0:              liq.CoinAddr,
			Token1:              liq.PcAddr,
			Amount0:             decimal.NewFromBigInt(new(big.Int).SetUint64(liq.InitCoinAmount), 0),
			Amount1:             decimal.NewFromBigInt(new(big.Int).SetUint64(liq.InitPcAmount), 0),
			Liquidity:           decimal.NewFromBigInt(new(big.Int).SetUint64(liq.Liq), 0),
			AddLiquidityAddress: liq.Authority,
			Block:               slot,
			Tx:                  tx,
			Dex:                 liq.Dex,
			Type:                pub.LIQ_CREATE,
			TxTime:              txTime.Unix(),
			ChainCode:           pub.SOLANA,
			CoinAccount:         liq.CoinAccount,
			PcAccount:           liq.PcAccount,
		}

		_, _, uPairs, nPairs := checkAndUpdatePair(txTime, MyHolds, tDecimalsMap, poolEvent, tx)
		if len(nPairs) > 0 {
			pair := nPairs[0]
			pair.Authority = poolEvent.AddLiquidityAddress
			newPairs = append(newPairs, pair)
		}
		poolE = append(poolE, poolEvent)
		updatePairs = append(updatePairs, uPairs...)

	}
	return poolE, updatePairs, newPairs
}

func checkQuote(event *model.PoolEvent) model.PoolEvent {
	quoteTokenIn, inOk := QuoteMap[event.Token0]
	quoteTokenOut, outOk := QuoteMap[event.Token1]
	// 两个都是报价币 另外一个不是USDC 则为其更新价格
	if inOk && outOk {
		if (quoteTokenIn.PairSymbol == "USDC") != (quoteTokenOut.PairSymbol == "USDC") {
			quoteTokenFlag := false
			if quoteTokenOut.PairSymbol == "USDC" {
				quoteTokenFlag = true
			}
			if !quoteTokenFlag {
				temp := event.Token0
				event.Token0 = event.Token1
				event.Token1 = temp
				temp1 := event.Amount0
				event.Amount0 = event.Amount1
				event.Amount1 = temp1
				temp2 := event.CoinAccount
				event.CoinAccount = event.PcAccount
				event.PcAccount = temp2
			}

		}

	} else if inOk && !outOk {
		temp := event.Token0
		event.Token0 = event.Token1
		event.Token1 = temp
		temp1 := event.Amount0
		event.Amount0 = event.Amount1
		event.Amount1 = temp1
		temp2 := event.CoinAccount
		event.CoinAccount = event.PcAccount
		event.PcAccount = temp2
	} else if !inOk && outOk {

	} else {

	}
	return *event
}

func checkAndGetPair(holds []model.TokenHold, solHolds map[string]uint64, sp pub.SwapInfo) (pairInCache *pub.Pair, exi bool, updatePairs []pub.Pair, newPairs []pub.Pair) {
	updatePairs = make([]pub.Pair, 0)
	newPairs = make([]pub.Pair, 0)
	MyHolds1 := make(map[string]decimal.Decimal)
	MyHoldsMint := make(map[string]string)
	for _, hold := range holds {
		MyHolds1[hold.Account] = hold.Amount
		MyHoldsMint[hold.Account] = hold.TokenAddress
	}
	coinAccount := sp.CoinAccount
	pcAccount := sp.PcAccount
	addr, ex := MyHoldsMint[pcAccount]

	fmt.Println(coinAccount)
	if ex {
		if addr == pub.SOL || addr == "SOL" {
			addr = pub.WSOL
		}
		if pub.DexMeteora != sp.Dex {
			if sp.QuoteToken != addr {
				coinAccount = sp.PcAccount
				pcAccount = sp.CoinAccount
			}
		} else {
			//DexMeteora 的虚拟池子,但池子的Token是对应的lp池子
			addr, ex := MyHoldsMint[sp.PcAccountTrue]
			if ex {
				if sp.QuoteToken != addr {
					coinAccount = sp.PcAccount
					pcAccount = sp.CoinAccount
				}
			}
		}

	}

	return pairInCache, false, updatePairs, newPairs
}

func SwapRaydium(slot uint64, txTime time.Time, swap *raydium.RaydiumLiqSwap, tokenDecimalsMap map[string]uint8, dex string, tx string) (sps []pub.SwapInfo) {
	defer func() {
		if r := recover(); r != nil {
			Log.Errorf("Swappanic: %s, %s ,error:%v", tx, swap.ToString(), r)
		}
	}()

	amountIn := swap.AmountIn
	amountOut := swap.AmountOut
	if amountIn < 0 || amountOut < 0 {
		Log.Errorf("Swappanic: tx:%s,   swap:%s  ", tx, swap.ToString())
	}
	//minimumAmountOut := swap.MinimumAmountOut
	pairAddr := swap.LpPair

	//authority := swap.Authority

	if swap.InAddr == nil || swap.OutAddr == nil {
		Log.Errorf("InAddr or OutAddr is nil %s,%s  ", tx, swap.ToString())
		return sps
	}
	inAddr := swap.InAddr
	outAddr := swap.OutAddr
	inAddrDecimals := tokenDecimalsMap[*inAddr]
	outAddrDecimals := tokenDecimalsMap[*outAddr]
	//主币deciaml为0 SOL设置9位小数
	if swap.Dex == pub.DexPumpFun || swap.Dex == pub.DexMoonshot {
		if outAddrDecimals == 0 {
			if *outAddr == pub.WSOL {
				outAddrDecimals = 9
			}
		}
		if inAddrDecimals == 0 {
			if *inAddr == pub.WSOL {
				inAddrDecimals = 9
			}
		}
	}
	if outAddrDecimals == 0 {
		q, inOk := QuoteMap[*outAddr]
		if inOk {
			outAddrDecimals = uint8(q.Decimals)
		}
	}
	if inAddrDecimals == 0 {
		q, inOk := QuoteMap[*inAddr]
		if inOk {
			inAddrDecimals = uint8(q.Decimals)
		}
	}
	uniInAmount := decimal.NewFromBigInt(new(big.Int).SetUint64(amountIn), 0).DivRound(decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(inAddrDecimals))), 18)
	uniOutAmount := decimal.NewFromBigInt(new(big.Int).SetUint64(amountOut), 0).DivRound(decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(outAddrDecimals))), 18)

	quoteTokenIn, inOk := QuoteMap[*inAddr]
	quoteTokenOut, outOk := QuoteMap[*outAddr]
	//两个都是报价币 其中至少有一个不为"USDC"(都为"USDC"则不解析)
	if inOk && outOk {
		// 两个都是报价币 另外一个不是USDC 则为其更新价格
		if (quoteTokenIn.PairSymbol == "USDC") != (quoteTokenOut.PairSymbol == "USDC") {
			baselastTime := quoteTokenOut.LestTime
			coinAmount := uniOutAmount
			quotAmount := uniInAmount
			base := outAddr
			basePrice := quoteTokenOut.Price
			quoteToken := quoteTokenIn
			baseDecimals := outAddrDecimals
			quoteDecimals := inAddrDecimals
			quoteTokenFlag := false
			if quoteTokenOut.PairSymbol == "USDC" {
				basePrice = quoteTokenIn.Price
				baselastTime = quoteTokenIn.LestTime
				base = inAddr
				coinAmount = uniInAmount
				quotAmount = uniOutAmount
				quoteToken = quoteTokenOut
				quoteTokenFlag = true
				baseDecimals = inAddrDecimals
				quoteDecimals = outAddrDecimals
			}
			fmt.Println(baselastTime, basePrice)
			//Log.Printf(fmt.Sprintf("ok ok symbol:%s price:%s ", quoteTokenOut.Symbol, quoteTokenOut.Price.String()))
			//Log.Println(fmt.Sprintf("ok ok symbol:%s price:%s ", quoteTokenIn.Symbol, quoteTokenIn.Price.String()))
			quotePrice := quoteToken.Price
			if quotePrice.Cmp(decimal.Zero) == 0 {
				quoteTokenOut := QuoteMap[*base]
				quotePrice = quoteTokenOut.Price
			}
			priceMutQuote := quotAmount.Mul(quotePrice)
			if coinAmount.Cmp(decimal.Zero) == 0 {
				Log.Infof("decimal division by 0 %s  , uniInAmount:%s, uniOutAmount:%s,price *MutQuote = %s*%s = %s  ", uniInAmount.String(), uniOutAmount.String(), quotePrice.String(), quotAmount, priceMutQuote, swap.ToString())
				return sps
			}
			price := priceMutQuote.DivRound(coinAmount, 18)
			//更新价格的前提是对应的pair地址
			if quoteTokenFlag {
				if *pairAddr == quoteTokenIn.PairAddress {
					QuoteMap[*inAddr].SetPrice(price)
					quoteTokenIn.Price = price
				}
			} else {
				if *pairAddr == quoteTokenOut.PairAddress {
					QuoteMap[*outAddr].SetPrice(price)
					quoteTokenOut.Price = price
				}
			}
			//Log.Println(fmt.Sprintf("ok ok symbol:%s price:%s ", quoteTokenOut.Symbol, quoteTokenOut.Price.String()))
			//Log.Println(fmt.Sprintf("ok ok symbol:%s price:%s ", quoteTokenIn.Symbol, quoteTokenIn.Price.String()))
			//Log.Println(fmt.Sprintf("ok ok in:%s out:%s ,usd:%s ,quotePrice:%s ,coinPrice:%s  ,pairAddr:%s ", uniInAmount, uniOutAmount, priceMutQuote, quotePrice, price, pairAddr))

			if price.Cmp(decimal.Zero) == 0 {
				Log.Infof("swaplog0 %s  , uniInAmount:%s, uniOutAmount:%s,price*quotAmount = %s*%s=%s %s ",
					tx, uniInAmount.String(), uniOutAmount.String(), quotePrice.String(), quotAmount.String(), priceMutQuote, swap.ToString())
				return sps
			}
			ct := swap.CoinAccountTrue
			pt := swap.PcAccountTrue
			if ct == nil {
				ct = swap.CoinAccount
			}
			if pt == nil {
				pt = swap.PcAccount
			}
			sps = append(sps, pub.SwapInfo{

				From:               *swap.WalletAddr,
				Pair:               *pairAddr,
				Dex:                swap.Dex,
				In:                 *inAddr,
				Out:                *outAddr,
				CoinAccount:        *swap.CoinAccount,
				PcAccount:          *swap.PcAccount,
				CoinAccountTrue:    *ct,
				PcAccountTrue:      *pt,
				AmountIn:           uniInAmount,
				AmountOut:          uniOutAmount,
				PriceMutQuotePrice: priceMutQuote,
				Price:              price,
				BaseToken:          *base,
				QuoteToken:         quoteToken.Address,
				CoinAmount:         coinAmount,
				QuoteAmount:        quotAmount,
				BaseDecimals:       baseDecimals,
				QuoteDecimals:      quoteDecimals,
				TxTime:             txTime.Unix(),
				TxHash:             tx,
				Slot:               slot,
				Curr:               time.Now().Unix(),
			})
		}

	} else if inOk && !outOk {
		coinAmount := uniOutAmount
		quotAmount := uniInAmount
		quoteToken := quoteTokenIn
		quotePrice := quoteToken.Price
		if quotePrice.Cmp(decimal.Zero) == 0 {
			quoteTokenOut := QuoteMap[*inAddr]
			quotePrice = quoteTokenOut.Price
		}
		priceMutQuote := quotAmount.Mul(quotePrice)
		if coinAmount.Cmp(decimal.Zero) == 0 {
			Log.Infof("swaplog0 %s  , uniInAmount:%s, uniOutAmount:%s,price*quotAmount = %s*%s=%s %s ",
				tx, uniInAmount.String(), uniOutAmount.String(), quotePrice.String(), quotAmount.String(), priceMutQuote, swap.ToString())
			return sps
		}
		price := priceMutQuote.DivRound(coinAmount, 30)
		if price.Cmp(decimal.Zero) == 0 {
			Log.Infof("swaplog0 %s  , uniInAmount:%s, uniOutAmount:%s,price*quotAmount = %s*%s=%s %s ",
				tx, uniInAmount.String(), uniOutAmount.String(), quotePrice.String(), quotAmount.String(), priceMutQuote, swap.ToString())
			return sps
		}
		ct := swap.CoinAccountTrue
		pt := swap.PcAccountTrue
		if ct == nil {
			ct = swap.CoinAccount
		}
		if pt == nil {
			pt = swap.PcAccount
		}
		//Log.Println(fmt.Sprintf("inOk && !outOk in:%s out:%s ,usd:%s ,%s:%s ,coinPrice:%s  ,pairAddr:%s ", uniInAmount, uniOutAmount, priceMutQuote, quoteToken.Symbol, quotePrice, price, pairAddr))
		sps = append(sps, pub.SwapInfo{
			From:               *swap.WalletAddr,
			Pair:               *pairAddr,
			Dex:                swap.Dex,
			In:                 *inAddr,
			Out:                *outAddr,
			CoinAccount:        *swap.CoinAccount,
			PcAccount:          *swap.PcAccount,
			CoinAccountTrue:    *ct,
			PcAccountTrue:      *pt,
			AmountIn:           uniInAmount,
			AmountOut:          uniOutAmount,
			PriceMutQuotePrice: priceMutQuote,
			Price:              price,
			BaseToken:          *outAddr,
			QuoteToken:         *inAddr,
			CoinAmount:         coinAmount,
			QuoteAmount:        quotAmount,
			BaseDecimals:       outAddrDecimals,
			QuoteDecimals:      inAddrDecimals,
			TxTime:             txTime.Unix(),
			TxHash:             tx,
			Slot:               slot,
			Curr:               time.Now().Unix(),
		})
	} else if !inOk && outOk {
		coinAmount := uniInAmount
		quotAmount := uniOutAmount
		quoteToken := quoteTokenOut

		quotePrice := quoteToken.Price
		if quotePrice.Cmp(decimal.Zero) == 0 {
			quoteTokenOut := QuoteMap[*outAddr]
			quotePrice = quoteTokenOut.Price
		}
		priceMutQuote := quotAmount.Mul(quotePrice)
		if coinAmount.Cmp(decimal.Zero) == 0 {
			Log.Infof("swaplog0 %s  , uniInAmount:%s, uniOutAmount:%s,price*quotAmount = %s*%s=%s %s ",
				tx, uniInAmount.String(), uniOutAmount.String(), quotePrice.String(), quotAmount.String(), priceMutQuote, swap.ToString())
			return sps
		}
		price := priceMutQuote.DivRound(coinAmount, 30)
		if price.Cmp(decimal.Zero) == 0 {
			Log.Infof("swaplog0 %s  , uniInAmount:%s, uniOutAmount:%s,price*quotAmount = %s*%s=%s %s ",
				tx, uniInAmount.String(), uniOutAmount.String(), quotePrice.String(), quotAmount.String(), priceMutQuote, swap.ToString())
			return sps
		}
		//Log.Println(fmt.Sprintf("!inOk && outOk in:%s out:%s ,usd:%s ,%s:%s ,coinPrice:%s ,pairAddr:%s ", uniInAmount, uniOutAmount, priceMutQuote, quoteToken.Symbol, quotePrice, price, pairAddr))
		ct := swap.CoinAccountTrue
		pt := swap.PcAccountTrue
		if ct == nil {
			ct = swap.CoinAccount
		}
		if pt == nil {
			pt = swap.PcAccount
		}
		sps = append(sps, pub.SwapInfo{
			From:               *swap.WalletAddr,
			Pair:               *pairAddr,
			Dex:                swap.Dex,
			In:                 *inAddr,
			Out:                *outAddr,
			CoinAccount:        *swap.CoinAccount,
			PcAccount:          *swap.PcAccount,
			CoinAccountTrue:    *ct,
			PcAccountTrue:      *pt,
			AmountIn:           uniInAmount,
			AmountOut:          uniOutAmount,
			PriceMutQuotePrice: priceMutQuote,
			Price:              price,
			BaseToken:          *inAddr,
			QuoteToken:         *outAddr,
			CoinAmount:         coinAmount,
			QuoteAmount:        quotAmount,
			BaseDecimals:       inAddrDecimals,
			QuoteDecimals:      outAddrDecimals,
			TxTime:             txTime.Unix(),
			TxHash:             tx,
			Slot:               slot,
			Curr:               time.Now().Unix(),
		})
	} else {
		//两个都是山寨币 则不解析
	}

	return sps
}

func processTempAccountMap(txSlot uint64, txTime time.Time, tokenDecimalsMap map[string]uint8, createAccounts []*DecodeInsData, accountMap map[string]pub.TempAccountData) {
	//收集临时账户
	for _, c := range createAccounts {
		key := c.Key
		accountIns := c.Data.(*token.Instruction)
		switch key {
		case pub.InitializeAccount:
			accountImpl := accountIns.Impl.(*token.InitializeAccount)
			account3 := accountImpl.AccountMetaSlice[0].PublicKey.String()
			_, exist := accountMap[account3]
			if !exist {
				mint := accountImpl.GetAccounts()[1].PublicKey.String()
				u, exis := tokenDecimalsMap[mint]
				aaccount := pub.TempAccountData{
					Account: account3,
					Mint:    accountImpl.GetMintAccount().PublicKey.String(),
					Owner:   accountImpl.GetOwnerAccount().PublicKey.String(),
					TxTime:  txTime,
					Block:   txSlot,
				}
				if exis {
					aaccount.Decimals = u
				}

				accountMap[account3] = aaccount
			}

		case pub.InitializeAccount2:
			accountImpl := accountIns.Impl.(*token.InitializeAccount2)
			account3 := accountImpl.AccountMetaSlice[0].PublicKey.String()
			_, exist := accountMap[account3]
			if !exist {
				mint := accountImpl.GetAccounts()[1].PublicKey.String()
				u, exis := tokenDecimalsMap[mint]
				aaccount := pub.TempAccountData{
					Account: account3,
					Mint:    accountImpl.GetMintAccount().PublicKey.String(),
					Owner:   accountImpl.Owner.String(),
					TxTime:  txTime,
					Block:   txSlot,
				}
				if exis {
					aaccount.Decimals = u
				}

				accountMap[account3] = aaccount
			}
		case pub.InitializeAccount3:
			accountImpl := accountIns.Impl.(*token.InitializeAccount3)
			account3 := accountImpl.AccountMetaSlice[0].PublicKey.String()
			_, exist := accountMap[account3]
			if !exist {
				mint := accountImpl.GetAccounts()[1].PublicKey.String()
				u, exis := tokenDecimalsMap[mint]
				aaccount := pub.TempAccountData{
					Account: account3,
					Mint:    accountImpl.GetMintAccount().PublicKey.String(),
					Owner:   accountImpl.Owner.String(),
					TxTime:  txTime,
					Block:   txSlot,
				}
				if exis {
					aaccount.Decimals = u
				}

				accountMap[account3] = aaccount
			}
		default:
			continue
		}
	}
}

func processTempAcountsAndBalance(txsl uint64, txTime time.Time, txMeta *rpc.TransactionMeta, metas solana.AccountMetaSlice, tokenDecimalsMap map[string]uint8, ths []model.TokenHold, accountMap map[string]pub.TempAccountData) (tokenhold []model.TokenHold, solHolds map[string]uint64, holdUpdateKeys []string, balanceChg []model.BalanceChange) {
	debug := false
	defer func() {
		if r := recover(); r != nil {
			if debug {
				buf := make([]byte, 10240)
				n := runtime.Stack(buf, true)
				Log.Errorf("PANIC: %v\nFull Stack Trace:\n%s", r, string(buf[:n]))
			}
		}
	}()

	balanceChg = make([]model.BalanceChange, 0)
	//spl hold,以account为纬度
	allHolds := make(map[string]model.TokenHold)
	//sol Decimals 总是为9
	tokenDecimalsMap[pub.SOLANA] = 9
	tokenDecimalsMap["SOL"] = 9
	tokenDecimalsMap["11111111111111111111111111111111"] = 9
	tokenDecimalsMap[pub.WSOL] = 9
	// 余额没有变化
	noChange := make(map[string]bool)

	//交易之前
	for _, balance := range txMeta.PreTokenBalances {
		sAccount := metas[balance.AccountIndex].PublicKey.String()
		_, exist := accountMap[sAccount]
		owner := balance.Owner.String()
		tokenAddr := balance.Mint.String()
		//组合spl account 和spltoken 和 owner 的映射
		if !exist {
			accountMap[sAccount] = pub.TempAccountData{
				Account: sAccount,
				Mint:    tokenAddr,
				Owner:   owner,
			}
		}

		decimals := balance.UiTokenAmount.Decimals
		tokenDecimalsMap[tokenAddr] = decimals

		allHolds[sAccount] = model.TokenHold{
			Address:      owner,
			TokenAddress: tokenAddr,
			WtaKey:       pub.HashKey(owner, tokenAddr),
			Amount:       decimal.RequireFromString(balance.UiTokenAmount.Amount),
			Decimals:     decimals,
			UpdateTime:   txTime,
			ChainCode:    "SOLANA",
			Account:      sAccount,
		}
	}
	//交易之后余额
	for i, balance := range txMeta.PostTokenBalances {
		sAccount := metas[balance.AccountIndex].PublicKey.String()
		_, exist := accountMap[sAccount]
		owner := balance.Owner.String()
		tokenAddr := balance.Mint.String()

		if !exist {
			accountMap[sAccount] = pub.TempAccountData{
				Account: sAccount,
				Mint:    tokenAddr,
				Owner:   owner,
			}
		}

		if i < len(txMeta.PreTokenBalances) {
			if txMeta.PreTokenBalances[i].UiTokenAmount.Amount == balance.UiTokenAmount.Amount {
				noChange[sAccount] = true

			}
		}
		Post := decimal.RequireFromString(balance.UiTokenAmount.Amount)
		Pre := decimal.Zero
		if i < len(txMeta.PreTokenBalances) {
			Pre = decimal.RequireFromString(txMeta.PreTokenBalances[i].UiTokenAmount.Amount)
		}
		Chg := Post.Sub(Pre).Abs()
		balanceChg = append(balanceChg, model.BalanceChange{
			Owner:    balance.Owner.String(),
			Token:    tokenAddr,
			Post:     Post,
			Pre:      Pre,
			Chg:      Chg,
			Decimals: int8(balance.UiTokenAmount.Decimals),
		})
		decimals := balance.UiTokenAmount.Decimals
		tokenDecimalsMap[tokenAddr] = decimals

		allHolds[sAccount] = model.TokenHold{
			Address:      owner,
			TokenAddress: tokenAddr,
			WtaKey:       pub.HashKey(owner, tokenAddr),
			Amount:       decimal.RequireFromString(balance.UiTokenAmount.Amount),
			Decimals:     decimals,
			UpdateTime:   txTime,
			ChainCode:    "SOLANA",
			Account:      sAccount,
		}
		// Moonshot accountMap 记录余额变化
		posBalance := decimal.RequireFromString(balance.UiTokenAmount.Amount)
		for _, preTokenBalance := range txMeta.PreTokenBalances {
			pAccount := metas[preTokenBalance.AccountIndex].PublicKey.String()
			if pAccount == sAccount {
				preBalance := decimal.RequireFromString(preTokenBalance.UiTokenAmount.Amount)
				abs := preBalance.Sub(posBalance).Abs()
				if abs.Cmp(decimal.Zero) > 0 {
					accountMap["chgs:"+sAccount] = pub.TempAccountData{
						Account: sAccount,
						Mint:    tokenAddr,
						Owner:   abs.String(),
					}
				}
			}
		}

	}
	//获取原生solana账户余额
	balancesPosts := txMeta.PostBalances
	balancesPres := txMeta.PreBalances

	if len(metas) >= len(balancesPosts) {
		solHolds = make(map[string]uint64)
		for i, postBalance := range balancesPosts {
			sAccount := metas[i].PublicKey.String()
			solHolds[sAccount] = postBalance

			//更新sol余额 有变动的更新
			if balancesPres[i] != postBalance {
				solhold := model.TokenHold{
					Address:      sAccount,
					TokenAddress: "SOL",
					WtaKey:       pub.HashKey(sAccount, "SOL"),
					Amount:       decimal.NewFromBigInt(new(big.Int).SetUint64(postBalance), 0),
					Decimals:     9,
					UpdateTime:   txTime,
					ChainCode:    "SOLANA",
					Account:      sAccount,
				}

				ths = append(ths, solhold)
			}
			Post := decimal.NewFromBigInt(new(big.Int).SetUint64(postBalance), 0)
			Pre := decimal.NewFromBigInt(new(big.Int).SetUint64(balancesPres[i]), 0)
			if i < len(txMeta.PreTokenBalances) {
				Pre = decimal.RequireFromString(txMeta.PreTokenBalances[i].UiTokenAmount.Amount)
			}
			Chg := Post.Sub(Pre).Abs()
			balanceChg = append(balanceChg, model.BalanceChange{
				Owner:    sAccount,
				Token:    "SOL",
				Pre:      Pre,
				Post:     Post,
				Chg:      Chg,
				Decimals: int8(9),
			})
			// todo Moonshot无法获取到sol转移数量 临时解决方案 在构建accountMap时将对应的account 余额填充进去
			res := new(big.Int).Sub(big.NewInt(0).SetUint64(postBalance), big.NewInt(0).SetUint64(balancesPres[i]))
			res = new(big.Int).Abs(res)
			if res.Cmp(big.NewInt(0)) != 0 {
				accountMap["chg:"+sAccount] = pub.TempAccountData{
					Account: sAccount,
					Mint:    pub.SOLANA,
					Owner:   res.String(),
				}
			}

		}
	}

	for s, data := range accountMap {
		data.TxTime = txTime
		data.Block = txsl

		u, exis := tokenDecimalsMap[data.Mint]
		if exis {
			data.Decimals = u
		}

		accountMap[s] = data
	}
	//传递区块信息
	accountMap[pub.SOLANA] = pub.TempAccountData{
		Account:  pub.SOLANA,
		Mint:     pub.SOLANA,
		Owner:    pub.SOLANA,
		Decimals: 9,
		TxTime:   txTime,
		Block:    txsl,
	}
	return ths, solHolds, holdUpdateKeys, balanceChg
}

func processTransferRecordEntry(tempData pub.DecodeInsTempData, transfers []*token.Instruction, sysTransfers []*system.Instruction, trChecked []*token.Instruction) (trs []model.TransferRecord) {
	accountMap := tempData.TempAccount
	tokenDecimalsMap := tempData.TokenDecimalsMap
	//遍历组装transfer
	for _, t := range transfers {
		transfer := t.Impl.(*token.Transfer)
		amount := transfer.Amount
		source := transfer.GetSourceAccount().PublicKey.String()
		destination := transfer.GetDestinationAccount().PublicKey.String()
		authority := transfer.GetOwnerAccount().PublicKey.String()
		tokenAddr := ""
		tempAccount, ok := accountMap[source]
		if ok {
			// 获取源账户信息 len>2为token.InitializeAccount mint地址为index[2]
			source = tempAccount.Owner
			tokenAddr = tempAccount.Mint
		} else {
			tempAccountBydestination, isOk := accountMap[destination]
			if isOk {
				destination = tempAccountBydestination.Owner
				tokenAddr = tempAccountBydestination.Mint
			} else {
				// 获取源账户信息
				sourceAccInfo, err := dego.GetAccountInfo(context.Background(), transfer.GetSourceAccount().PublicKey)
				if err != nil {
					Log.Printf("failed to get source account info: %v", err)
				} else {
					// 获取代币合约地址（mint 地址）
					if sourceAccInfo.Value == nil || sourceAccInfo.Value.Data == nil || len(sourceAccInfo.Value.Data.GetBinary()) < 32 {
						Log.Error("failed to get source account info+")
					}
					mintAddress := sourceAccInfo.Value.Data.GetBinary()[:32] // mint 地址通常是前 32 个字节
					tokenAddr = solana.PublicKeyFromBytes(mintAddress).String()

				}
			}

		}
		if tokenDecimals, ok := tokenDecimalsMap[tokenAddr]; ok {
			trs = append(trs, model.TransferRecord{
				Time:      tempData.TxTime,
				Tx:        tempData.Tx,
				From:      source,
				To:        destination,
				Value:     *amount,
				Block:     tempData.TxSlot,
				GasUsed:   tempData.Fee,
				TxTime:    tempData.TxTime.Unix(),
				Contract:  tokenAddr,
				ChainCode: "SOLANA",
				ChainID:   0,
				Authority: authority,
				Decimals:  tokenDecimals,
			})
		}
	}

	//遍历组装SysTransfer
	for _, s := range sysTransfers {
		transfer := s.Impl.(*system.Transfer)
		amount := transfer.Lamports
		from := transfer.AccountMetaSlice[0].PublicKey.String()
		to := transfer.AccountMetaSlice[1].PublicKey.String()
		//decimals := uint8(9)
		trs = append(trs, model.TransferRecord{
			Time:      tempData.TxTime,
			Tx:        tempData.Tx,
			From:      from,
			To:        to,
			Value:     *amount,
			Block:     tempData.TxSlot,
			GasUsed:   tempData.Fee,
			TxTime:    tempData.TxTime.Unix(),
			Contract:  "SOL",
			ChainCode: "SOLANA",
			ChainID:   0,
			Authority: "",
			Decimals:  9,
		})
		//spew.Dump(amount, from, to, decimals)
	}
	//遍历组装token.TransferChecked
	for _, s := range trChecked {
		transfer := s.Impl.(*token.TransferChecked)
		amount := transfer.Amount
		from := transfer.GetSourceAccount().PublicKey.String()
		to := transfer.GetDestinationAccount().PublicKey.String()
		//decimals := uint8(9)
		trs = append(trs, model.TransferRecord{
			Time:      tempData.TxTime,
			Tx:        tempData.Tx,
			From:      from,
			To:        to,
			Value:     *amount,
			Block:     tempData.TxSlot,
			GasUsed:   tempData.Fee,
			TxTime:    tempData.TxTime.Unix(),
			Contract:  transfer.GetMintAccount().PublicKey.String(),
			ChainCode: "SOLANA",
			ChainID:   0,
			Authority: transfer.GetOwnerAccount().PublicKey.String(),
			Decimals:  9,
		})
		//spew.Dump(amount, from, to, decimals)
	}
	return trs
}
