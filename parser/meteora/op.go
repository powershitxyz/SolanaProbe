package meteora

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/mr-tron/base58"
	"github.com/powershitxyz/SolanaProbe/dego"
	"github.com/powershitxyz/SolanaProbe/model"
	"github.com/powershitxyz/SolanaProbe/parser/raydium"
	"github.com/powershitxyz/SolanaProbe/pub"
)

type MeteoraDex struct {
	model.DexRouter
	Auth *string
}

func (r *MeteoraDex) ParseLiquidityCreate(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*model.ArchDexMod, error) {
	//allInner, ok := extra[0].(*[]solana.CompiledInstruction)
	//accountProgramKeysMeta, ok := extra[1].(solana.AccountMetaSlice)
	//var accountMap map[string]pub.TempAccountData
	//if len(extra) > 2 {
	//	if accountMapTmp, ok := extra[2].(map[string]pub.TempAccountData); ok {
	//		accountMap = accountMapTmp
	//	}
	//
	//}
	//if !ok {
	//	return nil, errors.New("type not match")
	//}

	decodedData := raydium.RaydiumLiqCreate{Dex: pub.DexMeteora}

	decodedData.LpPair = accounts[0].PublicKey.String()
	decodedData.Authority = accounts[8].PublicKey.String()
	decodedData.CoinAddr = accounts[2].PublicKey.String()
	decodedData.PcAddr = accounts[3].PublicKey.String()
	decodedData.CoinAccount = accounts[4].PublicKey.String()
	decodedData.PcAccount = accounts[5].PublicKey.String()
	decodedData.InitPcAmount = 0
	decodedData.InitPcAmount = 0
	//var count = 0
	// remove liquidity时 可能有两个transfer 也可能只有一个transfer
	//if allInner != nil {
	//	for _, si := range *allInner {
	//		if count >= 2 {
	//			continue
	//		}
	//		if accountProgramKeysMeta[si.ProgramIDIndex].PublicKey.String() == "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA" {
	//			trAccounts := make([]*solana.AccountMeta, 0)
	//			for _, v := range si.Accounts {
	//				trAccounts = append(trAccounts, accountProgramKeysMeta[v])
	//			}
	//			instru, err := token.DecodeInstruction(trAccounts, si.Data)
	//			if err != nil {
	//				continue
	//			}
	//			if instru.TypeID.Uint8() == 12 {
	//				count++
	//				transfer := instru.Impl.(*token.TransferChecked)
	//
	//				amount1 := transfer.Amount
	//				src := transfer.GetSourceAccount()
	//				dest := transfer.GetDestinationAccount()
	//				// auth := accounts[int(si.Accounts[2])]
	//				var tokenAddr *solana.PublicKey
	//				var queryKey solana.PublicKey
	//				//remove liquidity时 可能有两个transfer 也可能只有一个transfer
	//				//不确定remove 哪个代币
	//				if len(accountMap) > 0 {
	//					tmpData, exist := accountMap[src.PublicKey.String()]
	//					if exist {
	//						tokenAddrBytes, err := base58.Decode(tmpData.Mint)
	//						if err == nil {
	//							tokenAddrCV := solana.PublicKeyFromBytes(tokenAddrBytes[0:32])
	//							tokenAddr = &tokenAddrCV
	//						}
	//					} else if !exist {
	//						tmpData, exist := accountMap[dest.PublicKey.String()]
	//						if exist {
	//							tokenAddrBytes, err := base58.Decode(tmpData.Mint)
	//							if err == nil {
	//								tokenAddrCV := solana.PublicKeyFromBytes(tokenAddrBytes[0:32])
	//								tokenAddr = &tokenAddrCV
	//							}
	//						}
	//					} else {
	//						sourceAccInfo, err := dego.GetAccountInfo(context.Background(), queryKey)
	//						if err == nil {
	//							mintAddress := sourceAccInfo.Value.Data.GetBinary()[:32]
	//							out := solana.PublicKeyFromBytes(mintAddress)
	//							tokenAddr = &out
	//						}
	//					}
	//				}
	//				tokenCa := tokenAddr.String()
	//
	//
	//			}
	//		}
	//	}
	//}
	return &model.ArchDexMod{
		DexName:  pub.DexMeteora,
		TypeName: pub.LiqCreate,
		Data:     &decodedData,
	}, nil
}

func (r *MeteoraDex) ParseLiquidityRemove(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*model.ArchDexMod, error) {
	allInner, ok := extra[0].(*[]solana.CompiledInstruction)
	accountProgramKeysMeta, ok := extra[1].(solana.AccountMetaSlice)
	var accountMap map[string]pub.TempAccountData
	if len(extra) > 2 {
		if accountMapTmp, ok := extra[2].(map[string]pub.TempAccountData); ok {
			accountMap = accountMapTmp
		}

	}
	if !ok {
		return nil, errors.New("type not match")
	}

	decodedData := raydium.RaydiumLiqRemove{Dex: pub.DexMeteora}

	decodedData.LpPair = accounts[1].PublicKey.String()
	decodedData.Authority = accounts[11].PublicKey.String()
	decodedData.CoinAddr = accounts[7].PublicKey.String()
	decodedData.PcAddr = accounts[8].PublicKey.String()
	decodedData.CoinAccount = accounts[5].PublicKey.String()
	decodedData.PcAccount = accounts[6].PublicKey.String()
	decodedData.PcAmount = 0
	decodedData.CoinAmount = 0
	var count = 0
	// remove liquidity时 可能有两个transfer 也可能只有一个transfer
	if allInner != nil {
		for _, si := range *allInner {
			if count >= 2 {
				continue
			}
			programID := accountProgramKeysMeta[si.ProgramIDIndex].PublicKey.String()
			if programID == "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA" || "TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb" == programID {
				trAccounts := make([]*solana.AccountMeta, 0)
				for _, v := range si.Accounts {
					trAccounts = append(trAccounts, accountProgramKeysMeta[v])
				}
				instru, err := token.DecodeInstruction(trAccounts, si.Data)
				if err != nil {
					continue
				}
				if instru.TypeID.Uint8() == 12 {
					count++
					transfer := instru.Impl.(*token.TransferChecked)

					amount1 := transfer.Amount
					src := transfer.GetSourceAccount()
					dest := transfer.GetDestinationAccount()
					// auth := accounts[int(si.Accounts[2])]
					var tokenAddr *solana.PublicKey
					var queryKey solana.PublicKey
					//remove liquidity时 可能有两个transfer 也可能只有一个transfer
					//不确定remove 哪个代币
					if len(accountMap) > 0 {
						tmpData, exist := accountMap[src.PublicKey.String()]
						if exist {
							tokenAddrBytes, err := base58.Decode(tmpData.Mint)
							if err == nil {
								tokenAddrCV := solana.PublicKeyFromBytes(tokenAddrBytes[0:32])
								tokenAddr = &tokenAddrCV
							}
						} else if !exist {
							tmpData, exist := accountMap[dest.PublicKey.String()]
							if exist {
								tokenAddrBytes, err := base58.Decode(tmpData.Mint)
								if err == nil {
									tokenAddrCV := solana.PublicKeyFromBytes(tokenAddrBytes[0:32])
									tokenAddr = &tokenAddrCV
								}
							}
						} else {
							sourceAccInfo, err := dego.GetAccountInfo(context.Background(), queryKey)
							if err == nil {
								mintAddress := sourceAccInfo.Value.Data.GetBinary()[:32]
								out := solana.PublicKeyFromBytes(mintAddress)
								tokenAddr = &out
							}
						}
					}
					tokenCa := tokenAddr.String()
					if tokenCa == decodedData.CoinAddr {
						decodedData.PcAmount = *amount1
					}
					if tokenCa == decodedData.PcAddr {
						decodedData.CoinAmount = *amount1
					}

				}
			}
		}
	}
	return &model.ArchDexMod{
		DexName:  pub.DexMeteora,
		TypeName: pub.LiqRemove,
		Data:     &decodedData,
	}, nil
}

func (r *MeteoraDex) ParseLiquidityAdd(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*model.ArchDexMod, error) {
	allInner, ok := extra[0].(*[]solana.CompiledInstruction)
	accountProgramKeysMeta, ok := extra[1].(solana.AccountMetaSlice)
	//var accountMap map[string]pub.TempAccountData
	//if len(extra) > 2 {
	//	if accountMapTmp, ok := extra[2].(map[string]pub.TempAccountData); ok {
	//		accountMap = accountMapTmp
	//	}
	//
	//}
	if !ok {
		return nil, errors.New("type not match")
	}

	decodedData := raydium.RaydiumLiqAdd{Dex: pub.DexMeteora}

	decodedData.LpPair = accounts[1].PublicKey.String()
	decodedData.Authority = accounts[11].PublicKey.String()
	decodedData.CoinAddr = accounts[7].PublicKey.String()
	decodedData.PcAddr = accounts[8].PublicKey.String()
	decodedData.CoinAccount = accounts[5].PublicKey.String()
	decodedData.PcAccount = accounts[6].PublicKey.String()

	var count = 0
	if allInner != nil {
		for _, si := range *allInner {

			if accountProgramKeysMeta[si.ProgramIDIndex].PublicKey.String() == "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA" {
				trAccounts := make([]*solana.AccountMeta, 0)
				for _, v := range si.Accounts {
					trAccounts = append(trAccounts, accountProgramKeysMeta[v])
				}
				instru, err := token.DecodeInstruction(trAccounts, si.Data)
				if err != nil {
					continue
				}
				if instru.TypeID.Uint8() == 12 {
					count++
					tr := instru.Impl.(*token.TransferChecked)
					amount := tr.Amount
					if count == 1 {
						decodedData.MaxCoinAmount = *amount
					} else if count == 2 {
						decodedData.MaxPcAmount = *amount
					}
				}
			}
		}
	}
	return &model.ArchDexMod{
		DexName:  pub.DexMeteora,
		TypeName: pub.LiqAdd,
		Data:     &decodedData,
	}, nil
}

func (r *MeteoraDex) ParseLiquiditySwap(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*model.ArchDexMod, error) {
	if len(extra) < 2 {
		return nil, errors.New("wrong extra param length")
	}

	allInner, ok := extra[0].(*[]solana.CompiledInstruction)
	accountProgramKeysMeta, ok := extra[1].(solana.AccountMetaSlice)
	if !ok {
		return nil, errors.New("type not match")
	}

	//accountProgramKeysMeta, ok := extra[1].(solana.AccountMetaSlice)
	var accountMap map[string]pub.TempAccountData
	if len(extra) > 2 {
		if accountMapTmp, ok := extra[2].(map[string]pub.TempAccountData); ok {
			accountMap = accountMapTmp
		}

	}
	decodedData := raydium.RaydiumLiqSwap{Dex: pub.DexMeteora}

	lpStr := accounts[0].PublicKey.String()
	auth := accounts[10].PublicKey.String()
	decodedData.LpPair = &lpStr
	decodedData.Authority = &auth
	decodedData.WalletAddr = &auth
	ReserveX := accounts[2].PublicKey.String()
	ReserveY := accounts[3].PublicKey.String()
	tokenXMint := accounts[6].PublicKey.String()
	tokenYMint := accounts[7].PublicKey.String()

	decodedData.CoinAccount = &ReserveX
	decodedData.PcAccount = &ReserveY
	decodedData.CoinAccountTrue = &ReserveX
	decodedData.PcAccountTrue = &ReserveY
	decodedData.CoinAddr = &tokenXMint
	decodedData.PcAddr = &tokenYMint
	HostFeeIn := accounts[9].PublicKey.String()
	hasFeeAccount := HostFeeIn != "LBUZKhRxPF3XUpBCjp4YzTKgLccjZhTSDM9YuVaPwxo"
	var jupInn int
	flag := len(extra) > 4
	if flag {
		//获取索引  拿取allInner的  jupInn<index< jupInn+2
		jupIndex := extra[len(extra)-1]
		jupInn, ok = jupIndex.(int)
		if !ok {
			return nil, errors.New("type not match")
		}
	}

	var count = 0
	if allInner != nil {
		for index, si := range *allInner {

			if flag && (count > 3 || index > jupInn+4 || index <= jupInn) {
				continue
			}

			programID := accountProgramKeysMeta[si.ProgramIDIndex].PublicKey.String()
			if programID == "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA" || "TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb" == programID {

				trAccounts := make([]*solana.AccountMeta, 0)
				for _, v := range si.Accounts {
					trAccounts = append(trAccounts, accountProgramKeysMeta[v])
				}
				instru, err := token.DecodeInstruction(trAccounts, si.Data)
				if err != nil {
					continue
				}
				if instru.TypeID.Uint8() == 12 {
					count++
					transfer := instru.Impl.(*token.TransferChecked)
					amount := transfer.Amount
					src := transfer.GetSourceAccount()
					dest := transfer.GetDestinationAccount()
					//pub.Log.Info("src:", src.PublicKey.String(), "dest:", dest.PublicKey.String(), "amount:", amount)
					// auth := accounts[int(si.Accounts[2])]
					var tokenAddr *solana.PublicKey
					var queryKey solana.PublicKey
					if len(accountMap) > 0 {
						tmpData, exist := accountMap[src.PublicKey.String()]
						if exist {
							tokenAddrBytes, err := base58.Decode(tmpData.Mint)
							if err == nil {
								tokenAddrCV := solana.PublicKeyFromBytes(tokenAddrBytes[0:32])
								tokenAddr = &tokenAddrCV
							}
						} else if !exist {
							tmpData, exist := accountMap[dest.PublicKey.String()]
							if exist {
								tokenAddrBytes, err := base58.Decode(tmpData.Mint)
								if err == nil {
									tokenAddrCV := solana.PublicKeyFromBytes(tokenAddrBytes[0:32])
									tokenAddr = &tokenAddrCV
								}
							}
						} else {
							sourceAccInfo, err := dego.GetAccountInfo(context.Background(), queryKey)
							if err == nil {
								mintAddress := sourceAccInfo.Value.Data.GetBinary()[:32]
								out := solana.PublicKeyFromBytes(mintAddress)
								tokenAddr = &out
							}
						}
					}
					if !hasFeeAccount {
						if count == 1 {
							decodedData.AmountOut = *amount
							s := tokenAddr.String()
							decodedData.OutAddr = &s
						} else if count == 2 {
							decodedData.AmountIn = *amount
							s := tokenAddr.String()
							decodedData.InAddr = &s
						}
					} else {
						if count == 2 {
							decodedData.AmountOut = *amount
							s := tokenAddr.String()
							decodedData.OutAddr = &s
						} else if count == 3 {
							decodedData.AmountIn = *amount
							s := tokenAddr.String()
							decodedData.InAddr = &s
						}
					}

				}
			}
		}
	}

	return &model.ArchDexMod{
		DexName:  pub.DexMeteora,
		TypeName: pub.LiqSwap,
		Data:     &decodedData,
	}, nil
}

func (r *MeteoraDex) UniCall(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*model.ArchDexMod, error) {
	//defer func() {
	//	if r := recover(); r != nil {
	//		pub.Log.Errorf("MeteoraDexErr:  %s ,error:%v", extra, r)
	//	}
	//}()
	reader := bytes.NewReader(data)

	var discriminator byte
	if err := binary.Read(reader, binary.LittleEndian, &discriminator); err != nil {
		return nil, err
	}
	//tail -f /app/hellosol/data.log  | grep 'discriminator'
	// discriminator
	//248  Swap 2*TransferChecked
	// 228  Unknown

	//26  RemoveLiquidityByRange  1*TransferChecked  48fmF1hKE8NmeNAAB2rKg9a5QTyh4ZBYCoNgUkUnrkoqo4aJJLqPdcmAD1Ta4jQsLsLkcwiNrShXcCxBmqW9RDkv
	//169  ClaimFee
	//123  ClosePosition
	//48fmF1hKE8NmeNAAB2rKg9a5QTyh4ZBYCoNgUkUnrkoqo4aJJLqPdcmAD1Ta4jQsLsLkcwiNrShXcCxBmqW9RDkv
	//6g3tgj3uZMrwn6Up9nq4o6WyQjE2CMrM9vVPdAuCYsfru7v8JBiwqB6hJgcsc4nFoHQcLw8JQewCUN7LTpyjcxw

	//26  RemoveLiquidityByRange  2*TransferChecked  4DW2fuZACP2HVmeyUd94X9YvcbBSXhRJRDUuQ4R5HDZTs8hn7vVpHMGJeZvZVPdJrtVRy4UdJTEq4eprEKZMEPCj

	//创建
	//219  InitializePosition   4AXG87maKA1PApwGmc2WjAyJcnTQwwgwKpLCc9BfpYi2QZCszdHq6mLEzxqwYboHP7YbskrRaTecAuQWcA9mtyjq
	//7    AddLiquidityByStrategy

	//219  InitializePosition   5wu4ffu5Pg7PYWHWr8KEJ3EnuxZ14MosC4L5GVXWu8ndfHFTFCJpuNbw4M9zUTRevL2h4E4aCJzhwCwVcRc3N21S
	//35  InitializeBinArray
	//7    AddLiquidityByStrategy

	//7  AddLiquidityByStrategy  unknown+2*TransferChecked  2V3d2Ko5pLk54PyRwMSRYkQ6uHt2f6oTUh6Xo9D6D6QfaoEMDxopqHbf1SbLwBKimd3pLqn8QEXkrGxLeuWH8DWF

	// 148 GoToABin 2nejJUumbK8ZPzmNjcHp6SvRUs6BtAaXT9yuXucDDFemW4ziwLPa9AK1CTPhzXtQnJgxaEdf5u5z194ZBoiB7EC9

	// 45 InitializeLbPair 36SNia9jp6epGwrXJJhuAKVzw9ifzsrDi1Q8do8QxfaUYxFAeBUcm1EsoHPmZmEvddHwXw6xrSouygVjmuFWU44k
	switch discriminator {
	case 45:
		return r.ParseLiquidityCreate(accounts, data[1:], extra...)
	case 7:
		return r.ParseLiquidityAdd(accounts, data[1:], extra...)
	case 26:
		return r.ParseLiquidityRemove(accounts, data[1:], extra...)
	case 248:
		return r.ParseLiquiditySwap(accounts, data[1:], extra...)

	}

	return nil, fmt.Errorf("MeteoraDex no imple: %d %s", discriminator, extra[3].(string))
	// if discriminator == 1 {
	// 	return r.ParseLiquidityCreate(accounts, data[1:], extra...)
	// }
	// if discriminator == 4 {
	// 	return r.ParseLiquidityAdd(accounts, data[1:], extra...)
	// }

	// if len(data) == 9 {
	// 	return r.ParseLiquidityRemove(accounts, data, extra...)
	// }
	// if len(data) == 26 {
	// 	return r.ParseLiquidityAdd(accounts, data, extra...)
	// }
	// TODO Here need to parse swap
}
