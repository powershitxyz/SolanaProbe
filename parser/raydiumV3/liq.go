package raydiumV3

import (
	"encoding/json"
	"fmt"
	"math/big"
)

type RaydiumLiqCreate struct {
	Nonce          uint8
	OpenTime       uint64
	InitPcAmount   uint64
	InitCoinAmount uint64
	Liq            uint64

	Dex         string
	LpPair      string // 5
	Authority   string // 6
	LpAccount   string // 8
	CoinAddr    string // 9
	PcAddr      string // 10
	CoinAccount string
	PcAccount   string
}

type RaydiumLiqRemove struct {
	Liq         big.Int
	Dex         string
	Amount      uint64
	CoinAmount  uint64
	PcAmount    uint64
	LpPair      string // 1
	Authority   string // 2
	LpAccount   string // 5
	CoinAddr    string // 6
	PcAddr      string // 7
	CoinAccount string
	PcAccount   string
}

type RaydiumLiqAdd struct {
	Dex           string
	MaxCoinAmount uint64
	MaxPcAmount   uint64
	BaseSide      uint64
	Liq           big.Int

	LpPair      string // 1
	Authority   string // 2
	LpAccount   string // 5
	CoinAddr    string // 6
	PcAddr      string // 7
	CoinAccount string
	PcAccount   string
}

type RaydiumLiqSwap struct {
	Dex              string
	AmountIn         uint64
	MinimumAmountOut uint64
	AmountOut        uint64 // From Inner

	LpPair      *string // 1
	Authority   *string // 2
	CoinAddr    *string // 4
	PcAddr      *string // 5
	CoinAccount *string // 4
	PcAccount   *string
	WalletAddr  *string

	InAddr  *string
	OutAddr *string
}

// 实现 ToString 方法
func (r *RaydiumLiqSwap) ToString() string {

	marshal, err := json.Marshal(r)
	if err != nil {
		return fmt.Sprintf("RaydiumLiqSwap: %s", fmt.Sprintf(
			"Dex: %s\nAmountIn: %d\nMinimumAmountOut: %d\nAmountOut: %d\nLpPair: %s\nAuthority: %s\nCoinAddr: %s\nPcAddr: %s\nWalletAddr: %s\nInAddr: %s\nOutAddr: %s",
			r.Dex,
			r.AmountIn,
			r.MinimumAmountOut,
			r.AmountOut,
			stringOrNil(r.LpPair),
			stringOrNil(r.Authority),
			stringOrNil(r.CoinAddr),
			stringOrNil(r.PcAddr),
			stringOrNil(r.WalletAddr),
			stringOrNil(r.InAddr),
			stringOrNil(r.OutAddr),
		))
	}
	return fmt.Sprintf("RaydiumLiqSwap: %s", string(marshal))
}
func stringOrNil(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}
