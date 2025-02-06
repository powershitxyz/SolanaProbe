package model

import "github.com/gagliardetto/solana-go"

type ArchDexMod struct {
	DexName  string
	TypeName string

	Data interface{}
}

type DexRouter interface {
	UniCall(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*ArchDexMod, error)
	ParseLiquidityCreate(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*ArchDexMod, error)
	ParseLiquidityAdd(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*ArchDexMod, error)
	ParseLiquidityRemove(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*ArchDexMod, error)

	ParseLiquiditySwap(accounts []*solana.AccountMeta, data []byte, extra ...interface{}) (*ArchDexMod, error)
}
