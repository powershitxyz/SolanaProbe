package orca

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/mr-tron/base58"
	"github.com/powershitxyz/SolanaProbe/dego"
	"github.com/powershitxyz/SolanaProbe/model"
	"github.com/powershitxyz/SolanaProbe/parser/raydium"
	"github.com/powershitxyz/SolanaProbe/pub"
)

type OrcaDex struct {
	model.DexRouter
	Auth *string
}

func (r *OrcaDex) ParseLiquidityCreate(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*model.ArchDexMod, error) {
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

	decodedData := raydium.RaydiumLiqCreate{Dex: pub.DexOrca}

	decodedData.LpPair = accounts[6].PublicKey.String()
	decodedData.Authority = accounts[5].PublicKey.String()
	decodedData.CoinAddr = accounts[1].PublicKey.String()
	decodedData.PcAddr = accounts[2].PublicKey.String()
	decodedData.CoinAccount = accounts[7].PublicKey.String()
	decodedData.PcAccount = accounts[8].PublicKey.String()

	return &model.ArchDexMod{
		DexName:  pub.DexOrca,
		TypeName: pub.LIQ_CREATE,
		Data:     &decodedData,
	}, nil
}

func (r *OrcaDex) ParseLiquidityRemove(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*model.ArchDexMod, error) {
	_, ok := extra[0].(*[]solana.CompiledInstruction)
	//_, ok := extra[1].(solana.AccountMetaSlice)
	var accountMap map[string]pub.TempAccountData
	if len(extra) > 2 {
		if accountMapTmp, ok := extra[2].(map[string]pub.TempAccountData); ok {
			accountMap = accountMapTmp
		}

	}
	if !ok {
		return nil, errors.New("type not match")
	}

	decodedData := raydium.RaydiumLiqRemove{Dex: pub.DexOrca}
	reverseBytes := pub.ReverseBytes(data)
	reader := bytes.NewReader(reverseBytes)

	var liq [16]byte      //uint128
	var TokenMaxA [8]byte //uint64
	var TokenMaxB [8]byte //uint64
	if err := binary.Read(reader, binary.BigEndian, &TokenMaxB); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.BigEndian, &TokenMaxA); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.BigEndian, &liq); err != nil {
		return nil, err
	}
	b := new(big.Int)
	b.SetBytes(liq[:])
	decodedData.Liq = *b
	decodedData.CoinAmount = binary.BigEndian.Uint64(TokenMaxA[:])
	decodedData.PcAmount = binary.BigEndian.Uint64(TokenMaxB[:])

	decodedData.LpPair = accounts[0].PublicKey.String()
	decodedData.Authority = accounts[2].PublicKey.String()
	decodedData.CoinAccount = accounts[7].PublicKey.String()
	decodedData.PcAccount = accounts[8].PublicKey.String()
	tmpData, exist := accountMap[decodedData.CoinAccount]
	if exist {
		decodedData.CoinAddr = tmpData.Mint
	} else {
		TokenAddrAOwnerAccountA := accounts[5].PublicKey.String()
		tmpData, exist := accountMap[TokenAddrAOwnerAccountA]
		if exist {
			decodedData.CoinAddr = tmpData.Mint
		}
	}
	tmpData, exist = accountMap[decodedData.PcAccount]
	if exist {
		decodedData.PcAddr = tmpData.Mint
	} else {
		TokenAddrAOwnerAccountB := accounts[6].PublicKey.String()
		tmpData, exist := accountMap[TokenAddrAOwnerAccountB]
		if exist {
			decodedData.PcAddr = tmpData.Mint
		}
	}
	return &model.ArchDexMod{
		DexName:  "Orca",
		TypeName: "LiqRemove",
		Data:     &decodedData,
	}, nil
}

func (r *OrcaDex) ParseLiquidityAdd(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*model.ArchDexMod, error) {
	_, ok := extra[0].(*[]solana.CompiledInstruction)
	//_, ok := extra[1].(solana.AccountMetaSlice)
	var accountMap map[string]pub.TempAccountData
	if len(extra) > 2 {
		if accountMapTmp, ok := extra[2].(map[string]pub.TempAccountData); ok {
			accountMap = accountMapTmp
		}

	}
	if !ok {
		return nil, errors.New("type not match")
	}

	decodedData := raydium.RaydiumLiqAdd{Dex: pub.DexOrca}
	reverseBytes := pub.ReverseBytes(data)
	reader := bytes.NewReader(reverseBytes)

	var liq [16]byte      //uint128
	var TokenMaxA [8]byte //uint64
	var TokenMaxB [8]byte //uint64
	if err := binary.Read(reader, binary.BigEndian, &TokenMaxB); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.BigEndian, &TokenMaxA); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.BigEndian, &liq); err != nil {
		return nil, err
	}
	b := new(big.Int)
	b.SetBytes(liq[:])
	decodedData.Liq = *b
	decodedData.MaxCoinAmount = binary.BigEndian.Uint64(TokenMaxA[:])
	decodedData.MaxPcAmount = binary.BigEndian.Uint64(TokenMaxB[:])

	decodedData.LpPair = accounts[0].PublicKey.String()
	decodedData.Authority = accounts[2].PublicKey.String()
	decodedData.CoinAccount = accounts[7].PublicKey.String()
	decodedData.PcAccount = accounts[8].PublicKey.String()
	tmpData, exist := accountMap[decodedData.CoinAccount]
	if exist {
		decodedData.CoinAddr = tmpData.Mint
	} else {
		TokenAddrAOwnerAccountA := accounts[5].PublicKey.String()
		tmpData, exist := accountMap[TokenAddrAOwnerAccountA]
		if exist {
			decodedData.CoinAddr = tmpData.Mint
		}
	}
	tmpData, exist = accountMap[decodedData.PcAccount]
	if exist {
		decodedData.PcAddr = tmpData.Mint
	} else {
		TokenAddrAOwnerAccountB := accounts[6].PublicKey.String()
		tmpData, exist := accountMap[TokenAddrAOwnerAccountB]
		if exist {
			decodedData.PcAddr = tmpData.Mint
		}
	}

	return &model.ArchDexMod{
		DexName:  "Orca",
		TypeName: "LiqAdd",
		Data:     &decodedData,
	}, nil
}

func (r *OrcaDex) ParseLiquiditySwap(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*model.ArchDexMod, error) {
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
	decodedData := raydium.RaydiumLiqSwap{Dex: pub.DexOrca}
	reverseBytes := pub.ReverseBytes(data)
	reader := bytes.NewReader(reverseBytes)
	var Amount [8]byte
	var OtherAmountThreshold [8]byte
	var AmountSpecifiedIsInput [1]byte
	var SqrtPriceLimit [16]byte
	var AToB [1]byte
	if err := binary.Read(reader, binary.BigEndian, &AToB); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.BigEndian, &SqrtPriceLimit); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.BigEndian, &AmountSpecifiedIsInput); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.BigEndian, &OtherAmountThreshold); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.BigEndian, &Amount); err != nil {
		return nil, err
	}

	decodedData.AmountIn = binary.BigEndian.Uint64(Amount[:])

	AToBVal := AToB[0] != 0
	lpStr := accounts[2].PublicKey.String()
	auth := accounts[1].PublicKey.String()
	decodedData.LpPair = &lpStr
	decodedData.Authority = &auth
	decodedData.WalletAddr = &auth
	TokenVaultA := accounts[4].PublicKey.String()
	TokenVaultB := accounts[6].PublicKey.String()
	OwnerAccountA := accounts[3].PublicKey.String()
	OwnerAccountB := accounts[5].PublicKey.String()
	tmpData, exist := accountMap[TokenVaultA]
	decodedData.CoinAccount = &TokenVaultB
	decodedData.PcAccount = &TokenVaultA
	decodedData.CoinAccountTrue = &TokenVaultB
	decodedData.PcAccountTrue = &TokenVaultA
	if exist {
		addrStr := tmpData.Mint
		decodedData.OutAddr = &addrStr
		decodedData.CoinAddr = &addrStr

	} else {
		tmpData, exist := accountMap[OwnerAccountA]
		if exist {
			addrStr := tmpData.Mint
			decodedData.OutAddr = &addrStr
			decodedData.CoinAddr = &addrStr
		}
	}
	tmpData1, exist := accountMap[TokenVaultB]
	if exist {
		addrStr := tmpData1.Mint
		decodedData.InAddr = &addrStr
		decodedData.PcAddr = &addrStr
	} else {
		tmpData1, exist := accountMap[OwnerAccountB]
		if exist {
			addrStr := tmpData1.Mint
			decodedData.InAddr = &addrStr
			decodedData.PcAddr = &addrStr

		}
	}
	if !AToBVal {
		temp := decodedData.InAddr
		decodedData.InAddr = decodedData.OutAddr
		decodedData.OutAddr = temp
	}
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

			if flag && (count > 2 || index > jupInn+2 || index <= jupInn) {
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
				if instru.TypeID.Uint8() == 3 {
					count++
					transfer := instru.Impl.(*token.Transfer)
					amount := transfer.Amount
					if count == 1 {
						decodedData.AmountOut = *amount
					} else if count == 2 {
						decodedData.AmountIn = *amount
					} else {
						continue
					}

				}
			}
		}
	}

	if decodedData.AmountIn == 0 || decodedData.AmountOut == 0 {
		pub.Log.Errorf("swaplog0:Orca ,len(extra):%d 聚合:%v ,%d ,%s  ", len(extra), flag, count, extra[3])
		return nil, errors.New("sb orca amount=0 ")
	}
	return &model.ArchDexMod{
		DexName:  "Orca",
		TypeName: "LiqSwap",
		Data:     &decodedData,
	}, nil
}
func (r *OrcaDex) ParseLiquiditySwapV2(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*model.ArchDexMod, error) {
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
	decodedData := raydium.RaydiumLiqSwap{Dex: pub.DexOrca}
	reverseBytes := pub.ReverseBytes(data)
	reader := bytes.NewReader(reverseBytes)

	var Amount [8]byte
	var OtherAmountThreshold [8]byte
	var SqrtPriceLimit [16]byte
	var AmountSpecifiedIsInput [1]byte
	var AToB [1]byte
	if err := binary.Read(reader, binary.BigEndian, &AToB); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.BigEndian, &AToB); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.BigEndian, &SqrtPriceLimit); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.BigEndian, &AmountSpecifiedIsInput); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.BigEndian, &OtherAmountThreshold); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.BigEndian, &Amount); err != nil {
		return nil, err
	}

	decodedData.AmountOut = binary.BigEndian.Uint64(Amount[:])

	AToBVal := AToB[0] != 0
	lpStr := accounts[4].PublicKey.String()
	auth := accounts[3].PublicKey.String()
	decodedData.LpPair = &lpStr
	decodedData.Authority = &auth
	decodedData.WalletAddr = &auth
	TokenVaultA := accounts[8].PublicKey.String()
	TokenVaultB := accounts[10].PublicKey.String()
	OwnerAccountA := accounts[7].PublicKey.String()
	OwnerAccountB := accounts[9].PublicKey.String()
	tmpData, exist := accountMap[TokenVaultA]
	decodedData.CoinAccount = &TokenVaultB
	decodedData.PcAccount = &TokenVaultA
	decodedData.CoinAccountTrue = &TokenVaultB
	decodedData.PcAccountTrue = &TokenVaultA
	if exist {
		addrStr := tmpData.Mint
		decodedData.OutAddr = &addrStr
		decodedData.CoinAddr = &addrStr

	} else {
		tmpData, exist := accountMap[OwnerAccountA]
		if exist {
			addrStr := tmpData.Mint
			decodedData.OutAddr = &addrStr
			decodedData.CoinAddr = &addrStr
		}
	}
	tmpData1, exist := accountMap[TokenVaultB]
	if exist {
		addrStr := tmpData1.Mint
		decodedData.InAddr = &addrStr
		decodedData.PcAddr = &addrStr
	} else {
		tmpData1, exist := accountMap[OwnerAccountB]
		if exist {
			addrStr := tmpData1.Mint
			decodedData.InAddr = &addrStr
			decodedData.PcAddr = &addrStr

		}
	}
	if !AToBVal {
		temp := decodedData.InAddr
		decodedData.InAddr = decodedData.OutAddr
		decodedData.OutAddr = temp
	}
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

			if flag && (count > 2 || index > jupInn+4 || index <= jupInn) {
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
				if instru.TypeID.Uint8() == 3 {
					count++
					transfer := instru.Impl.(*token.Transfer)
					amount := transfer.Amount
					src := transfer.GetSourceAccount()
					dest := transfer.GetDestinationAccount()
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
					if count == 1 {
						mint1 := tokenAddr.String()
						decodedData.OutAddr = &mint1
						decodedData.AmountOut = *amount
					} else if count == 2 {
						mint2 := tokenAddr.String()
						decodedData.InAddr = &mint2
						decodedData.AmountIn = *amount
					}

				}
				if instru.TypeID.Uint8() == 12 {
					count++
					transfer := instru.Impl.(*token.TransferChecked)
					amount1 := transfer.Amount
					src := transfer.GetSourceAccount()
					dest := transfer.GetDestinationAccount()
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
					if count == 1 {
						mint1 := tokenAddr.String()
						decodedData.OutAddr = &mint1
						decodedData.AmountOut = *amount1
					} else if count == 2 {
						mint2 := tokenAddr.String()
						decodedData.InAddr = &mint2
						decodedData.AmountIn = *amount1
					} else {
						continue
					}

				}
			}
			if count >= 2 {
				break
			}
		}
	}

	if decodedData.AmountIn == 0 || decodedData.AmountOut == 0 {
		pub.Log.Errorf("swaplog0:OrcaV2 ,len(extra):%d 聚合:%v ,%d ,%s  ", len(extra), flag, count, extra[3])
		return nil, errors.New("sb orca amount=0 ")
	}
	return &model.ArchDexMod{
		DexName:  "Orca",
		TypeName: "LiqSwap",
		Data:     &decodedData,
	}, nil
}

func (r *OrcaDex) UniCall(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*model.ArchDexMod, error) {
	defer func() {
		if r := recover(); r != nil {
			pub.Log.Errorf("OrcaDexErr:  %s ,error:%v", extra[3], r)
		}
	}()
	reader := bytes.NewReader(data)

	var discriminator byte
	if err := binary.Read(reader, binary.LittleEndian, &discriminator); err != nil {
		return nil, err
	}

	//242 OpenPositionWithMetadata //46 AddLiquidity
	//160 Remove //164 CollectFees  //248 swap

	switch discriminator {
	case 207:
		return r.ParseLiquidityCreate(accounts, data[1:], extra...)
	case 46:
		return r.ParseLiquidityAdd(accounts, data[1:], extra...)
	case 160:
		return r.ParseLiquidityRemove(accounts, data[1:], extra...)
	case 248:
		return r.ParseLiquiditySwap(accounts, data[1:], extra...)
	case 43:
		return r.ParseLiquiditySwapV2(accounts, data[1:], extra...)

	}

	return nil, fmt.Errorf("OrcaDex no imple: %d %s", discriminator, extra[3].(string))
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
