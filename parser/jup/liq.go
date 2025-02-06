package jup

import "github.com/powershitxyz/SolanaProbe/parser/raydium"

type JupSwap struct {
	Data   []*raydium.RaydiumLiqSwap
	Source string
}
