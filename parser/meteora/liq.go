package meteora

type MeteoraLiqCreate struct {
	InitPcAmount   uint64
	InitCoinAmount uint64
	Liq            uint64
	Dex            string
	LpPair         string //0
	Authority      string //8
	LpAccount      string
	CoinAddr       string //2
	PcAddr         string //3
	CoinAccount    string //4
	PcAccount      string //5
}

type RaydiumLiqRemove struct {
	Amount      uint64
	CoinAmount  uint64
	PcAmount    uint64
	LpPair      string // 1
	Authority   string // 11
	LpAccount   string //
	CoinAddr    string // 7
	PcAddr      string // 8
	CoinAccount string //5
	PcAccount   string //6
}

type MeteoraLiqAdd struct {
	MaxCoinAmount uint64
	MaxPcAmount   uint64

	LpPair      string // 1
	Authority   string // 11
	LpAccount   string //
	CoinAddr    string // 7
	PcAddr      string // 8
	CoinAccount string //5
	PcAccount   string //6
}
type MeteoraLiqSwap struct {
	/*{
	  	"amountIn": {
	  	"type": "u64",
	  	"data": " ?"
	  },
	  	"minAmountOut": {
	  	"type": "u64",
	  	"data": "0"
	  }
	  }*/
	AmountIn            uint64
	AmountOut           uint64
	LpPair              string // 0
	TokenXMint          string //6
	TokenYMint          string //7
	ReserveX            string //2
	ReserveY            string //3
	UserTokenAccountIn  string //4
	UserTokenAccountOut string //5
	User                string // 10
}
