package raydiumV3

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"

	"github.com/davecgh/go-spew/spew"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/powershitxyz/SolanaProbe/dego"
	"github.com/powershitxyz/SolanaProbe/model"
	"github.com/powershitxyz/SolanaProbe/parser/raydium"
	"github.com/powershitxyz/SolanaProbe/pub"
)

type RaydiumV3Dex struct {
	model.DexRouter
	Auth *string
}

func (r *RaydiumV3Dex) ParseLiquidityCreate(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*model.ArchDexMod, error) {

	return &model.ArchDexMod{
		DexName:  "Raydium",
		TypeName: "LiqCreate",
		Data:     nil,
	}, nil
}

func (r *RaydiumV3Dex) ParseLiquidityRemove(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*model.ArchDexMod, error) {
	allInner, ok := extra[0].(*[]solana.CompiledInstruction)
	accountProgramKeysMeta, ok := extra[1].(solana.AccountMetaSlice)

	if !ok {
		return nil, errors.New("type not match")
	}
	if len(accounts) < 16 {
		return nil, errors.New("account length miss match")
	}

	decodedData := raydium.RaydiumLiqRemove{Dex: pub.DexRaydium}
	reader := bytes.NewReader(data)

	var liquidity [16]byte
	var ligU128 big.Int

	if err := binary.Read(reader, binary.LittleEndian, &liquidity); err != nil {
		return nil, err
	}
	ligU128.SetBytes(liquidity[:])
	decodedData.Liq = ligU128

	decodedData.LpPair = accounts[3].PublicKey.String()
	decodedData.Authority = accounts[0].PublicKey.String()
	decodedData.LpAccount = accounts[1].PublicKey.String()
	decodedData.CoinAddr = accounts[14].PublicKey.String()
	decodedData.PcAddr = accounts[15].PublicKey.String()
	decodedData.CoinAccount = accounts[5].PublicKey.String()
	decodedData.PcAccount = accounts[6].PublicKey.String()
	if allInner != nil {
		for _, si := range *allInner {

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
					transfer := instru.Impl.(*token.TransferChecked)

					amount1 := transfer.Amount
					src := transfer.GetSourceAccount()
					dest := transfer.GetDestinationAccount()
					if src.PublicKey.String() == decodedData.CoinAccount || dest.PublicKey.String() == decodedData.CoinAccount {

						decodedData.CoinAmount = *amount1
					}
					if src.PublicKey.String() == decodedData.PcAccount || dest.PublicKey.String() == decodedData.PcAccount {

						decodedData.PcAmount = *amount1
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

func (r *RaydiumV3Dex) ParseLiquidityAdd(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*model.ArchDexMod, error) {
	allInner, ok := extra[0].(*[]solana.CompiledInstruction)
	accountProgramKeysMeta, ok := extra[1].(solana.AccountMetaSlice)

	if !ok {
		return nil, errors.New("type not match")
	}
	if len(accounts) < 15 {
		return nil, errors.New("account length miss match")
	}
	decodedData := raydium.RaydiumLiqAdd{Dex: pub.DexRaydium}
	reader := bytes.NewReader(data)

	var liquidity [16]byte
	var ligU128 big.Int

	if err := binary.Read(reader, binary.LittleEndian, &liquidity); err != nil {
		return nil, err
	}
	ligU128.SetBytes(liquidity[:])
	decodedData.Liq = ligU128
	decodedData.Authority = accounts[0].PublicKey.String()
	decodedData.LpAccount = accounts[1].PublicKey.String()
	decodedData.LpPair = accounts[2].PublicKey.String()
	decodedData.CoinAddr = accounts[13].PublicKey.String()
	decodedData.PcAddr = accounts[14].PublicKey.String()
	decodedData.CoinAccount = accounts[9].PublicKey.String()
	decodedData.PcAccount = accounts[10].PublicKey.String()

	if allInner != nil {
		for _, si := range *allInner {

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
					transfer := instru.Impl.(*token.TransferChecked)
					amount1 := transfer.Amount
					src := transfer.GetSourceAccount()
					dest := transfer.GetDestinationAccount()
					if src.PublicKey.String() == decodedData.CoinAccount || dest.PublicKey.String() == decodedData.CoinAccount {
						decodedData.MaxCoinAmount = *amount1
					}
					if src.PublicKey.String() == decodedData.PcAccount || dest.PublicKey.String() == decodedData.PcAccount {
						decodedData.MaxPcAmount = *amount1
					}

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

func (r *RaydiumV3Dex) ParseLiquiditySwap(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*model.ArchDexMod, error) {
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
	decodedData := raydium.RaydiumLiqSwap{Dex: pub.DexRaydium}
	//reverseBytes := pub.ReverseBytes(data)

	lpStr := accounts[2].PublicKey.String()
	auth := accounts[0].PublicKey.String()
	pcAccount := accounts[5].PublicKey.String()
	coinAccount := accounts[6].PublicKey.String()
	InAccount := coinAccount
	OutAccount := pcAccount

	decodedData.LpPair = &lpStr
	decodedData.Authority = &auth

	decodedData.PcAccount = &pcAccount
	decodedData.CoinAccount = &coinAccount
	decodedData.PcAccountTrue = &pcAccount
	decodedData.CoinAccountTrue = &coinAccount
	decodedData.WalletAddr = decodedData.Authority
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
				var amount *uint64
				var src, dest string
				if instru.TypeID.Uint8() == 3 {
					count++
					transfer := instru.Impl.(*token.Transfer)

					amount = transfer.Amount
					src = transfer.GetSourceAccount().PublicKey.String()
					dest = transfer.GetDestinationAccount().PublicKey.String()
				}
				if instru.TypeID.Uint8() == 12 {
					count++
					transfer := instru.Impl.(*token.TransferChecked)

					amount = transfer.Amount
					src = transfer.GetSourceAccount().PublicKey.String()
					dest = transfer.GetDestinationAccount().PublicKey.String()
				}

				if src == InAccount || dest == InAccount {
					if len(accountMap) > 0 {
						tmpData, exist := accountMap[src]
						var mint string
						if exist {
							mint = tmpData.Mint
						} else if !exist {
							tmpData, exist = accountMap[dest]
							if exist {
								mint = tmpData.Mint
							}
						} else {
							sourceAccInfo, err := dego.GetAccountInfo(context.Background(), solana.PublicKeyFromBytes([]byte(src)))
							if err == nil {
								mintAddress := sourceAccInfo.Value.Data.GetBinary()[:32]
								out := solana.PublicKeyFromBytes(mintAddress)
								mint = out.String()
							}
						}
						decodedData.AmountIn = *amount
						decodedData.InAddr = &mint
						decodedData.CoinAddr = &mint
					}
				}

				if src == OutAccount || dest == OutAccount {
					if len(accountMap) > 0 {
						tmpData, exist := accountMap[src]
						var mint string
						if exist {
							mint = tmpData.Mint
						} else if !exist {
							tmpData, exist = accountMap[dest]
							if exist {
								mint = tmpData.Mint
							}
						} else {
							sourceAccInfo, err := dego.GetAccountInfo(context.Background(), solana.PublicKeyFromBytes([]byte(src)))
							if err == nil {
								mintAddress := sourceAccInfo.Value.Data.GetBinary()[:32]
								out := solana.PublicKeyFromBytes(mintAddress)
								mint = out.String()
							}
						}
						decodedData.AmountOut = *amount
						decodedData.OutAddr = &mint
						decodedData.PcAddr = &mint
					}
				}

			}

		}
	}
	if decodedData.AmountIn == 0 || decodedData.AmountOut == 0 {
		spew.Dump("swaplog0 Raydium extra", extra[0], extra[3])
		fmt.Printf("swaplog0:Raydium ,len(extra):%d %v ,%d  \n", len(extra), flag, count)
	}
	return &model.ArchDexMod{
		DexName:  "Raydium",
		TypeName: "LiqSwap",
		Data:     &decodedData,
	}, nil
}

func (r *RaydiumV3Dex) UniCall(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*model.ArchDexMod, error) {
	defer func() {
		if r := recover(); r != nil {
			pub.Log.Errorf("RaydiumV3DexErr:  %s ,error:%v", extra[3], r)
		}
	}()
	reader := bytes.NewReader(data)

	var discriminator byte
	if err := binary.Read(reader, binary.LittleEndian, &discriminator); err != nil {
		return nil, err
	}
	//swap 4zW5coNjLjewFAvCurTfnRTUneygnMxTkzGgpoaba9DWnQANw5QhpZWiCpCGYLyy7U1DX4Z4REpjRG5EHTAkkTyh
	//swapv2 3v8DRXJU1KAjv3foMH9PRGxTYCgxcGUBBwbSTUZJx7HRTpoiTaqha2aBkwdJC87J2XZjBAPWJ1J3NrEmeF94dVhw
	//remove 2qKR2UU5hQu1iA18snYZr3kvr1fTN1D7jVf7RE9jVPxG4p4mUj4iSRwDJ6QzbxCtecr1NmjbER3fN7MmzExNE9Zw
	//add 5mBU3PF37en9zhvBS5xzP9CnLH1bRvDprhZnA6cQ5f2aVQCdyGaJb93tpcgPM8ywBjUZeSv5scF8kn6KFjw74BxR
	//log.Printf("v3: %v accounts: %d,%s", discriminator, len(accounts), extra[3])
	switch discriminator {
	case 133:
		return r.ParseLiquidityAdd(accounts, data[1:], extra...)
	case 58:
		return r.ParseLiquidityRemove(accounts, data[1:], extra...)
	case 43:
		//SWAP
		return r.ParseLiquiditySwap(accounts, data[1:], extra...)
	case 248:
		//SWAPv2
		return r.ParseLiquiditySwap(accounts, data[1:], extra...)

	}

	return nil, fmt.Errorf("RaydiumDexDex no imple: %d %s", discriminator, extra[3].(string))

}
