package jup

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/gagliardetto/solana-go"

	"github.com/powershitxyz/SolanaProbe/model"
	"github.com/powershitxyz/SolanaProbe/parser/raydium"
	"github.com/powershitxyz/SolanaProbe/parser/raydiumV3"
	"github.com/powershitxyz/SolanaProbe/pub"
)

type JupDex struct {
	model.DexRouter
}

func (r *JupDex) ParseLiquidityCreate(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*model.ArchDexMod, error) {
	return nil, errors.New("unsupport op")
}

func (r *JupDex) ParseLiquidityRemove(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*model.ArchDexMod, error) {
	return nil, errors.New("unsupport op")
}

func (r *JupDex) ParseLiquidityAdd(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*model.ArchDexMod, error) {
	return nil, errors.New("unsupport op")
}

func (r *JupDex) ParseLiquiditySwap(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*model.ArchDexMod, error) {
	if len(extra) < 2 {
		return nil, errors.New("wrong extra param length")
	}

	allInner, ok := extra[0].(*[]solana.CompiledInstruction)
	accountProgramKeysMeta, ok := extra[1].(solana.AccountMetaSlice)
	if !ok {
		return nil, errors.New("type not match")
	}

	accountProgramKeys, ok := extra[1].([]solana.PublicKey)
	if !ok {
		accountProgramKeysMeta, ok := extra[1].(solana.AccountMetaSlice)
		if !ok {
			return nil, errors.New("type not match3")
		}
		accountProgramKeys = append(accountProgramKeys, accountProgramKeysMeta.GetKeys()...)
	}
	authAddr := accounts[2].PublicKey.String()
	decodedData := &JupSwap{}

	innerDatas := make([]*raydium.RaydiumLiqSwap, 0)
	if allInner != nil {
		for index, si := range *allInner {
			programKeys := accountProgramKeysMeta[si.ProgramIDIndex].PublicKey.String()
			if programKeys == "675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8" {
				//Raydium
				trAccounts := make([]*solana.AccountMeta, 0)
				for _, v := range si.Accounts {
					trAccounts = append(trAccounts, accountProgramKeysMeta[v])
				}
				dex := &raydium.RaydiumDex{Auth: &authAddr}

				extra = append(extra, index)
				call, err := dex.UniCall(trAccounts, si.Data, extra...)
				if err == nil && call != nil && call.TypeName == pub.LiqSwap {
					innerDatas = append(innerDatas, call.Data.(*raydium.RaydiumLiqSwap))
				}
			} else if programKeys == "CAMMCzo5YL8w4VFF8KVHrK22GGUsp5VTaW7grrKgrWqK" {
				//Raydiumv3
				trAccounts := make([]*solana.AccountMeta, 0)
				for _, v := range si.Accounts {
					trAccounts = append(trAccounts, accountProgramKeysMeta[v])
				}
				dex := &raydiumV3.RaydiumV3Dex{Auth: &authAddr}

				extra = append(extra, index)
				call, err := dex.UniCall(trAccounts, si.Data, extra...)

				if err == nil && call != nil && call.TypeName == pub.LiqSwap {
					innerDatas = append(innerDatas, call.Data.(*raydium.RaydiumLiqSwap))
				}

			}
			// else if programKeys == "whirLbMiicVdio4qvUfM5KAg6Ct8VwpYzGff3uctyCc" {
			// 	//OrcaV2
			// 	trAccounts := make([]*solana.AccountMeta, 0)
			// 	for _, v := range si.Accounts {
			// 		trAccounts = append(trAccounts, accountProgramKeysMeta[v])
			// 	}
			// 	dex := &orca.OrcaDex{Auth: &authAddr}

			// 	extra = append(extra, index)
			// 	call, err := dex.UniCall(trAccounts, si.Data, extra...)

			// 	if err == nil && call != nil && call.TypeName == pub.LiqSwap {
			// 		innerDatas = append(innerDatas, call.Data.(*raydium.RaydiumLiqSwap))
			// 	}

			// } else if programKeys == "LBUZKhRxPF3XUpBCjp4YzTKgLccjZhTSDM9YuVaPwxo" {
			// 	//Meteora
			// 	trAccounts := make([]*solana.AccountMeta, 0)
			// 	for _, v := range si.Accounts {
			// 		trAccounts = append(trAccounts, accountProgramKeysMeta[v])
			// 	}
			// 	dex := &meteora.MeteoraDex{Auth: &authAddr}

			// 	extra = append(extra, index)
			// 	call, err := dex.UniCall(trAccounts, si.Data, extra...)

			// 	if err == nil {
			// 		innerDatas = append(innerDatas, call.Data.(*raydium.RaydiumLiqSwap))
			// 	}

			// } else if programKeys == "6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P" {
			// 	//PumpFun
			// 	trAccounts := make([]*solana.AccountMeta, 0)
			// 	for _, v := range si.Accounts {
			// 		trAccounts = append(trAccounts, accountProgramKeysMeta[v])
			// 	}
			// 	dex := &pumpfun.PumpFunDex{Auth: &authAddr}

			// 	extra = append(extra, index)
			// 	call, err := dex.UniCall(trAccounts, si.Data, extra...)

			// 	if err == nil && call != nil && call.TypeName == pub.LiqSwap {
			// 		innerDatas = append(innerDatas, call.Data.(*raydium.RaydiumLiqSwap))
			// 	}

			// } else if programKeys == "MoonCVVNZFSYkqNXP6bxHLPL6QQJiMagDL3qcqUQTrG" {
			// 	//Moonshot
			// 	trAccounts := make([]*solana.AccountMeta, 0)
			// 	for _, v := range si.Accounts {
			// 		trAccounts = append(trAccounts, accountProgramKeysMeta[v])
			// 	}
			// 	dex := &moonshot.MoonShotDex{Auth: &authAddr}

			// 	extra = append(extra, index)
			// 	call, err := dex.UniCall(trAccounts, si.Data, extra...)

			// 	if err == nil && call != nil && call.TypeName == pub.LiqSwap {
			// 		innerDatas = append(innerDatas, call.Data.(*raydium.RaydiumLiqSwap))

			// 	}

			// } else if programKeys == "Eo7WjKq67rjJQSZxS6z3YkapzY3eMj6Xy8X5EQVn5UaB" {
			// 	//MeteoraPools
			// 	trAccounts := make([]*solana.AccountMeta, 0)
			// 	for _, v := range si.Accounts {
			// 		trAccounts = append(trAccounts, accountProgramKeysMeta[v])
			// 	}
			// 	dex := &meteoraPools.MeteoraPoolsDex{Auth: &authAddr}

			// 	extra = append(extra, index)
			// 	call, err := dex.UniCall(trAccounts, si.Data, extra...)

			// 	if err == nil && call != nil && call.TypeName == pub.LiqSwap {
			// 		innerDatas = append(innerDatas, call.Data.(*raydium.RaydiumLiqSwap))

			// 	}

			// } else if programKeys == "5quBtoiQqxF9Jv6KYKctB59NT3gtJD2Y65kdnB1Uev3h" {
			// 	//MeteoraPools
			// 	trAccounts := make([]*solana.AccountMeta, 0)
			// 	for _, v := range si.Accounts {
			// 		trAccounts = append(trAccounts, accountProgramKeysMeta[v])
			// 	}
			// 	dex := &raydiumAmm.RaydiumAmmDex{Auth: &authAddr}

			// 	extra = append(extra, index)
			// 	call, err := dex.UniCall(trAccounts, si.Data, extra...)

			// 	if err == nil && call != nil && call.TypeName == pub.LiqSwap {
			// 		innerDatas = append(innerDatas, call.Data.(*raydium.RaydiumLiqSwap))

			// 	}

			// } else if programKeys == "CPMMoo8L3F4NbTegBCKVNunggL7H1ZpdTHKxQB5qKP1C" {
			// 	//MeteoraPools
			// 	trAccounts := make([]*solana.AccountMeta, 0)
			// 	for _, v := range si.Accounts {
			// 		trAccounts = append(trAccounts, accountProgramKeysMeta[v])
			// 	}
			// 	dex := &raydiumCpmm.RaydiumCPMMDex{Auth: &authAddr}

			// 	extra = append(extra, index)
			// 	call, err := dex.UniCall(trAccounts, si.Data, extra...)

			// 	if err == nil && call != nil && call.TypeName == pub.LiqSwap {
			// 		innerDatas = append(innerDatas, call.Data.(*raydium.RaydiumLiqSwap))

			// 	}

		}
	}

	//if len(ids) > 0 {
	//	vs := make([]*raydium.RaydiumLiqSwap, 0, len(ids))
	//	for _, v := range ids {
	//		vs = append(vs, innerData[v])
	//	}
	//	decodedData.Data = vs
	//}
	decodedData.Data = append(decodedData.Data, innerDatas...)
	return &model.ArchDexMod{
		DexName:  "JupiterV6",
		TypeName: "LiqSwap",
		Data:     decodedData,
	}, nil
}

func (r *JupDex) UniCall(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*model.ArchDexMod, error) {
	defer func() {
		if r := recover(); r != nil {
			pub.Log.Errorf("JupDexErr:  %s ,error:%v", extra[3], r)
		}
	}()
	reader := bytes.NewReader(data)

	var discriminator byte
	if err := binary.Read(reader, binary.LittleEndian, &discriminator); err != nil {
		return nil, err
	}
	//// ExactOutRoute 208
	switch discriminator {
	case 193:
		return r.ParseLiquiditySwap(accounts, data[1:], extra...)
	case 229:
		return r.ParseLiquiditySwap(accounts, data[1:], extra...)

	}

	return nil, fmt.Errorf("JupDex no imple: %d %s", discriminator, extra[3].(string))
}
