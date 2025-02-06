package raydium

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
	Dex              string `json:"dex"`
	AmountIn         uint64 `json:"amountIn"`
	MinimumAmountOut uint64 `json:"minimumAmountOut"`
	AmountOut        uint64 `json:"amountOut"`

	LpPair          *string `json:"lpPair"`
	Authority       *string `json:"authority"`
	CoinAddr        *string `json:"coinAddr"`
	PcAddr          *string `json:"pcAddr"`
	CoinAccount     *string `json:"coinAccount"`
	PcAccount       *string `json:"pcAccount"`
	CoinAccountTrue *string `json:"coinAccountTrue"`
	PcAccountTrue   *string `json:"pcAccountTrue"`
	WalletAddr      *string `json:"walletAddr"`

	InAddr  *string `json:"inAddr"`
	OutAddr *string `json:"outAddr"`
}

// 实现 ToString 方法

func (t *RaydiumLiqSwap) ToString() string {

	marshal, err := json.Marshal(t)
	if err != nil {
		return fmt.Sprintf("Swap: %s", fmt.Sprintf("%+v", t))

	}
	return fmt.Sprintf("Swap: %s", string(marshal))
}
func stringOrNil(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}
