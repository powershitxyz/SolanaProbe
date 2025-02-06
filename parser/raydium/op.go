package raydium

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/log"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/mr-tron/base58"
	"github.com/powershitxyz/SolanaProbe/dego"
	"github.com/powershitxyz/SolanaProbe/model"
	"github.com/powershitxyz/SolanaProbe/pub"
)

type RaydiumDex struct {
	model.DexRouter
	Auth *string
}

func (r *RaydiumDex) ParseLiquidityCreate(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*model.ArchDexMod, error) {
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

	decodedData := RaydiumLiqCreate{Dex: pub.DexRaydium}
	reader := bytes.NewReader(data)

	var nonce byte
	var openTime, initPcAmount, initCoinAmount [8]byte

	if err := binary.Read(reader, binary.LittleEndian, &nonce); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &openTime); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &initPcAmount); err != nil {
		return nil, err
	}

	if err := binary.Read(reader, binary.LittleEndian, &initCoinAmount); err != nil {
		return nil, err
	}

	decodedData.Nonce = nonce
	decodedData.OpenTime = binary.LittleEndian.Uint64(openTime[:])
	decodedData.InitPcAmount = binary.LittleEndian.Uint64(initPcAmount[:])
	decodedData.InitCoinAmount = binary.LittleEndian.Uint64(initCoinAmount[:])

	decodedData.LpPair = accounts[4].PublicKey.String()
	decodedData.Authority = accounts[17].PublicKey.String()
	decodedData.LpAccount = accounts[7].PublicKey.String()
	decodedData.CoinAddr = accounts[8].PublicKey.String()
	decodedData.PcAddr = accounts[9].PublicKey.String()
	decodedData.PcAccount = accounts[11].PublicKey.String()
	decodedData.CoinAccount = accounts[10].PublicKey.String()
	var token0 = ""
	var token1 = ""
	data0, exit := accountMap[decodedData.CoinAddr]
	data1, exit1 := accountMap[decodedData.PcAddr]
	if exit {
		token0 = data0.Mint
		decodedData.CoinAddr = token0
	}
	if exit1 {
		token1 = data1.Mint
		decodedData.PcAddr = token1
	}
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
				if instru.TypeID.Uint8() == 7 {
					mintTo := instru.Impl.(*token.MintTo)
					decodedData.Liq = *mintTo.Amount

				}
			}
		}
	}
	return &model.ArchDexMod{
		DexName:  "Raydium",
		TypeName: "LiqCreate",
		Data:     &decodedData,
	}, nil
}

func (r *RaydiumDex) ParseLiquidityRemove(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*model.ArchDexMod, error) {
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

	decodedData := RaydiumLiqRemove{Dex: pub.DexRaydium}
	reader := bytes.NewReader(data)

	var amount [8]byte

	if err := binary.Read(reader, binary.LittleEndian, &amount); err != nil {
		return nil, err
	}

	decodedData.Amount = binary.LittleEndian.Uint64(amount[:])

	decodedData.LpPair = accounts[1].PublicKey.String()
	decodedData.Authority = accounts[18].PublicKey.String()
	decodedData.LpAccount = accounts[5].PublicKey.String()
	decodedData.CoinAddr = accounts[6].PublicKey.String()
	decodedData.PcAddr = accounts[7].PublicKey.String()
	var token0 = ""
	var token1 = ""
	data0, exit := accountMap[decodedData.CoinAddr]
	data1, exit1 := accountMap[decodedData.PcAddr]
	if exit {
		token0 = data0.Mint
		decodedData.CoinAddr = token0
	}
	if exit1 {
		token1 = data1.Mint
		decodedData.PcAddr = token1
	}
	if allInner != nil {
		for index, si := range *allInner {

			if accountProgramKeysMeta[si.ProgramIDIndex].PublicKey.String() == "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA" {
				trAccounts := make([]*solana.AccountMeta, 0)
				for _, v := range si.Accounts {
					trAccounts = append(trAccounts, accountProgramKeysMeta[v])
				}
				instru, err := token.DecodeInstruction(trAccounts, si.Data)
				if err != nil {
					continue
				}
				if instru.TypeID.Uint8() == 3 {
					transfer := instru.Impl.(*token.Transfer)

					amount1 := transfer.Amount
					src := transfer.GetSourceAccount()
					dest := transfer.GetDestinationAccount()
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
							queryKey = dest.PublicKey
						} else {
							sourceAccInfo, err := dego.GetAccountInfo(context.Background(), queryKey)
							if err == nil {
								mintAddress := sourceAccInfo.Value.Data.GetBinary()[:32]
								out := solana.PublicKeyFromBytes(mintAddress)
								tokenAddr = &out
							}
						}
					}

					if index == 0 {
						decodedData.CoinAddr = tokenAddr.String()
						decodedData.CoinAmount = *amount1
					} else if index == 1 {
						decodedData.PcAddr = tokenAddr.String()
						decodedData.PcAmount = *amount1
					} else {
						continue
					}

				}
			}
		}
	}
	return &model.ArchDexMod{
		DexName:  "Raydium",
		TypeName: "LiqRemove",
		Data:     &decodedData,
	}, nil
}

func (r *RaydiumDex) ParseLiquidityAdd(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*model.ArchDexMod, error) {
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

	decodedData := RaydiumLiqAdd{Dex: pub.DexRaydium}
	reader := bytes.NewReader(data)

	var maxCoinAmount [8]byte //uint64
	var maxPcAmount [8]byte   //uint64
	var baseSide [8]byte      //uint64

	if err := binary.Read(reader, binary.LittleEndian, &maxCoinAmount); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &maxPcAmount); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &baseSide); err != nil {
		return nil, err
	}

	decodedData.MaxCoinAmount = binary.LittleEndian.Uint64(maxCoinAmount[:]) //uint64
	decodedData.MaxPcAmount = binary.LittleEndian.Uint64(maxPcAmount[:])     //uint64
	decodedData.BaseSide = binary.LittleEndian.Uint64(baseSide[:])           //uint64

	decodedData.LpPair = accounts[1].PublicKey.String()
	decodedData.Authority = accounts[12].PublicKey.String()
	decodedData.LpAccount = accounts[5].PublicKey.String()
	decodedData.CoinAddr = accounts[6].PublicKey.String()
	decodedData.PcAddr = accounts[7].PublicKey.String()
	var token0 = ""
	var token1 = ""
	data0, exit := accountMap[decodedData.CoinAddr]
	data1, exit1 := accountMap[decodedData.PcAddr]
	if exit {
		token0 = data0.Mint
		decodedData.CoinAddr = token0
	}
	if exit1 {
		token1 = data1.Mint
		decodedData.PcAddr = token1
	}

	if allInner != nil {
		for index, si := range *allInner {

			if accountProgramKeysMeta[si.ProgramIDIndex].PublicKey.String() == "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA" && index == 2 {
				trAccounts := make([]*solana.AccountMeta, 0)
				for _, v := range si.Accounts {
					trAccounts = append(trAccounts, accountProgramKeysMeta[v])
				}
				instru, err := token.DecodeInstruction(trAccounts, si.Data)
				if err != nil {
					continue
				}
				if instru.TypeID.Uint8() == 7 {
					mintTo := instru.Impl.(*token.MintTo)
					b := new(big.Int)
					b.SetUint64(*mintTo.Amount)
					decodedData.Liq = *b

				}
			}
		}
	}
	return &model.ArchDexMod{
		DexName:  "Raydium",
		TypeName: "LiqAdd",
		Data:     &decodedData,
	}, nil
}

func (r *RaydiumDex) ParseLiquiditySwap(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*model.ArchDexMod, error) {
	if len(extra) < 2 {
		return nil, errors.New("wrong extra param length")
	}

	allInner, ok := extra[0].(*[]solana.CompiledInstruction)

	if !ok {
		return nil, errors.New("type not match")
	}

	accountProgramKeysMeta, ok := extra[1].(solana.AccountMetaSlice)
	var accountMap map[string]pub.TempAccountData
	if len(extra) > 2 {
		if accountMapTmp, ok := extra[2].(map[string]pub.TempAccountData); ok {
			accountMap = accountMapTmp
		}

	}
	decodedData := RaydiumLiqSwap{Dex: pub.DexRaydium}
	reader := bytes.NewReader(data)

	var amountIn [8]byte
	var minimumAmountOut [8]byte
	// var amountOut [8]byte

	if err := binary.Read(reader, binary.LittleEndian, &amountIn); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &minimumAmountOut); err != nil {
		return nil, err
	}

	decodedData.AmountOut = binary.LittleEndian.Uint64(amountIn[:])
	decodedData.MinimumAmountOut = binary.LittleEndian.Uint64(minimumAmountOut[:])
	// decodedData.AmountOut = binary.LittleEndian.Uint64(amountOut[:])
	lpStr := accounts[1].PublicKey.String()
	auth := accounts[2].PublicKey.String()
	decodedData.LpPair = &lpStr
	decodedData.Authority = &auth
	index17 := 4
	if len(accounts) == 18 {
		index17 = index17 + 1
	}
	coinAccount := accounts[index17].PublicKey.String()
	pcAccount := accounts[index17+1].PublicKey.String()
	tmpCoinData, exist := accountMap[coinAccount]
	if exist {
		coinStr1 := tmpCoinData.Mint

		decodedData.CoinAddr = &coinStr1
		decodedData.CoinAccountTrue = &coinAccount
		decodedData.CoinAccount = &coinAccount
	}

	tmpCoinData, exist = accountMap[pcAccount]
	if exist {
		pcStr1 := tmpCoinData.Mint
		decodedData.PcAddr = &pcStr1
		decodedData.PcAccount = &pcAccount
		decodedData.PcAccountTrue = &pcAccount

	}

	userAddr := accounts[len(accounts)-1].PublicKey.String()
	decodedData.WalletAddr = &userAddr
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

			if accountProgramKeysMeta[si.ProgramIDIndex].PublicKey.String() == "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA" {
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
					if count == 1 {
						ta := tokenAddr.String()
						decodedData.OutAddr = &ta
					} else if count == 2 {
						ta := tokenAddr.String()
						decodedData.InAddr = &ta
						decodedData.AmountIn = *amount
					} else {
						continue
					}

				}
			}
		}
	}
	if decodedData.AmountIn == 0 || decodedData.AmountOut == 0 {
		log.Error("swaplog0 Raydium extra", extra[0], extra[3])
		fmt.Printf("swaplog0:Raydium ,len(extra):%d %v ,%d  \n", len(extra), flag, count)
	}
	return &model.ArchDexMod{
		DexName:  "Raydium",
		TypeName: "LiqSwap",
		Data:     &decodedData,
	}, nil
}

func (r *RaydiumDex) UniCall(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*model.ArchDexMod, error) {
	defer func() {
		if r := recover(); r != nil {
			pub.Log.Errorf("RaydiumDexErr:  %s ,error:%v", extra[3], r)
		}
	}()
	reader := bytes.NewReader(data)

	var discriminator byte
	if err := binary.Read(reader, binary.LittleEndian, &discriminator); err != nil {
		return nil, err
	}

	switch discriminator {
	case 1:
		return r.ParseLiquidityCreate(accounts, data[1:], extra...)
	case 3:
		return r.ParseLiquidityAdd(accounts, data[1:], extra...)
	case 4:
		return r.ParseLiquidityRemove(accounts, data[1:], extra...)
	case 9:
		return r.ParseLiquiditySwap(accounts, data[1:], extra...)

	}

	return nil, fmt.Errorf("RaydiumDex no imple: %d %s", discriminator, extra[3].(string))
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
