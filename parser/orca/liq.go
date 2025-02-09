package orca

import (
	"math/big"
)

type OrcaLiqCreate struct {
	Nonce          uint8
	OpenTime       uint64
	InitPcAmount   uint64
	InitCoinAmount uint64
	Liq            uint64

	LpPair      string // 6
	Authority   string // 5
	LpAccount   string //
	CoinAddr    string // 1
	PcAddr      string // 2
	CoinAccount string //7
	PcAccount   string //8
}

type OrcaLiqRemove struct {
	Liq     big.Int
	AmountA uint64
	AmountB uint64

	LpPair             string // 0
	Authority          string // 2
	TokenVaultAccountA string // 7
	TokenVaultAccountB string // 8
	TokenAddrA         string
	TokenAddrB         string
}

type OrcaLiqAdd struct {
	Liq     big.Int
	AmountA uint64 //AmountA
	AmountB uint64

	LpPair             string // 0
	Authority          string // 2
	TokenVaultAccountA string // 7
	TokenVaultAccountB string // 8
	TokenAddrA         string
	TokenAddrB         string
}

type OrcaLiqTwoHopSwap struct {
	/* {
		"amount": {
		"type": "u64",
		"data": "1000000"
	},
		"otherAmountThreshold": {
		"type": "u64",
		"data": "20313056"
	},
		"amountSpecifiedIsInput": {
		"type": "bool",
		"data": true
	},
		"aToBOne": {
		"type": "bool",
		"data": true
	},
		"aToBTwo": {
		"type": "bool",
		"data": false
	},
		"sqrtPriceLimitOne": {
		"type": "u128",
		"data": "4295048016"
	},
		"sqrtPriceLimitTwo": {
		"type": "u128",
		"data": "79226673515401279992447579055"
	}
	}*/
	Amount                 uint64
	OtherAmountThreshold   uint64
	AmountSpecifiedIsInput bool
	AToB1                  bool
	AToB2                  bool
	SqrtPriceLimit1        big.Int
	SqrtPriceLimit2        big.Int

	Amount1A uint64
	Amount1B uint64
	Amount2A uint64
	Amount2B uint64

	Authority string // 1

	LpPair1             string //2
	LpPair2             string //3
	TokenVault1A        string //5
	TokenVault1B        string //7
	TokenVault2A        string //9
	TokenVault2B        string //11
	TokenVaultAccount1A string //4
	TokenVaultAccount1B string //6
	TokenVaultAccount2A string //8
	TokenVaultAccount2B string //10
	TokenAddr1A         string //
	TokenAddr1B         string //
	TokenAddr2A         string //
	TokenAddr2B         string //
}

type OrcaLiqSwap struct {
	/*{
	  	"amount": {
	  	"type": "u64",
	  	"data": "10000000"
	  },
	  	"otherAmountThreshold": {
	  	"type": "u64",
	  	"data": "633783"
	  },
	  	"sqrtPriceLimit": {
	  	"type": "u128",
	  	"data": "4295048016"
	  },
	  	"amountSpecifiedIsInput": {
	  	"type": "bool",
	  	"data": true
	  },
	  	"aToB": {
	  	"type": "bool",
	  	"data": true
	  }
	  }*/
	Amount                 uint64
	OtherAmountThreshold   uint64
	AmountSpecifiedIsInput bool
	SqrtPriceLimit         big.Int
	AToB                   bool

	AmountA uint64
	AmountB uint64

	Authority     string // 1
	LpPair        string // 2
	TokenVaultA   string // 4
	TokenVaultB   string // 6
	OwnerAccountA string // 3
	OwnerAccountB string // 5
	TokenAddrA    string //
	TokenAddrB    string //
}
