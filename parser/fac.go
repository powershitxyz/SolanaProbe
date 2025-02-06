package parser

import (
	"errors"

	"github.com/powershitxyz/SolanaProbe/model"
	"github.com/powershitxyz/SolanaProbe/parser/jup"
	"github.com/powershitxyz/SolanaProbe/parser/raydium"
	"github.com/powershitxyz/SolanaProbe/parser/raydiumAmm"
	raydiumCpmm "github.com/powershitxyz/SolanaProbe/parser/raydiumCPMM"
	"github.com/powershitxyz/SolanaProbe/parser/raydiumV3"
)

type DexType string

const (
	Raydium      DexType = "675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8"
	Raydiumv3    DexType = "CAMMCzo5YL8w4VFF8KVHrK22GGUsp5VTaW7grrKgrWqK"
	RaydiumAmm   DexType = "5quBtoiQqxF9Jv6KYKctB59NT3gtJD2Y65kdnB1Uev3h"
	RaydiumCPMM  DexType = "CPMMoo8L3F4NbTegBCKVNunggL7H1ZpdTHKxQB5qKP1C"
	JupiterV6    DexType = "JUP6LkbZbjS1jKKwapdHNy74zcZ3tLUZoi5QNyVTaV4"
	OrcaV2       DexType = "whirLbMiicVdio4qvUfM5KAg6Ct8VwpYzGff3uctyCc"
	Meteora      DexType = "LBUZKhRxPF3XUpBCjp4YzTKgLccjZhTSDM9YuVaPwxo"
	MeteoraPools DexType = "Eo7WjKq67rjJQSZxS6z3YkapzY3eMj6Xy8X5EQVn5UaB"
	FluxBeam     DexType = "FLUXubRmkEi2q6K3Y9kBPg9248ggaZVsoSFhtJHSrm1X"
	PumpFun      DexType = "6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P"
	Moonshot     DexType = "MoonCVVNZFSYkqNXP6bxHLPL6QQJiMagDL3qcqUQTrG"
	OkxProxy     DexType = "6m2CDdhRgxpH4WjvdzxAYbGxwdGUz5MziiL5jek2kBma"
	Other        DexType = "Other"
)

func NewDexRouter(dexType DexType) (model.DexRouter, error) {
	switch dexType {
	case Raydium:
		return &raydium.RaydiumDex{}, nil
	case Raydiumv3:
		return &raydiumV3.RaydiumV3Dex{}, nil
	case RaydiumCPMM:
		return &raydiumCpmm.RaydiumCPMMDex{}, nil
	case RaydiumAmm:
		return &raydiumAmm.RaydiumAmmDex{}, nil
	case JupiterV6:
		return &jup.JupDex{}, nil
	// case OrcaV2:
	// 	return &orca.OrcaDex{}, nil
	// case Meteora:
	// 	return &meteora.MeteoraDex{}, nil
	// case MeteoraPools:
	// 	return &meteoraPools.MeteoraPoolsDex{}, nil
	// case PumpFun:
	// 	return &pumpfun.PumpFunDex{}, nil
	// case Moonshot:
	// 	return &moonshot.MoonShotDex{}, nil
	// case Other:
	// 	return &other.OtherDex{}, nil
	// case OkxProxy:
	// 	return &okxproxy.OkxProxyDex{}, nil
	default:
		return nil, errors.New("unsupport type")
	}
}
