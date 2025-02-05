package model

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

type Token struct {
	ID           uint64  `gorm:"column:id;primaryKey;autoIncrement;not null" json:"id"`
	Name         string  `gorm:"column:name;type:varchar;not null" json:"name"`
	Symbol       string  `gorm:"column:symbol;type:varchar;not null" json:"symbol"`
	Address      string  `gorm:"column:address;type:varchar;not null" json:"address"`
	Decimals     int     `gorm:"column:decimals;type:int;not null" json:"decimals"`
	TotalSupply  uint64  `gorm:"column:total_supply;type:numeric;not null" json:"total_supply"`
	Block        int64   `gorm:"column:block;type:bigint" json:"block"`
	Tx           string  `gorm:"column:tx;type:varchar" json:"tx"`
	Deploy       string  `gorm:"column:deploy;type:varchar" json:"deploy"`
	CreateAt     int64   `gorm:"column:create_at;type:bigint" json:"create_at"`
	ChainID      int64   `gorm:"column:chain_id;type:bigint;default:0" json:"chain_id"`
	OpeningQuote float64 `gorm:"column:opening_quote;type:numeric;default:0" json:"opening_quote"`
	Circulation  float64 `gorm:"column:circulation;type:numeric" json:"circulation"`
	Holders      int64   `gorm:"column:holders;type:bigint" json:"holders"`
	ChainCode    string  `gorm:"column:chain_code;type:varchar" json:"chain_code"`
	MetaUri      string  `gorm:"column:meta_uri;type:varchar" json:"meta_uri"`
}

// SolTokenHold 对应数据库中的 sol_token_holds 表

type TokenHold struct {
	ID            uint64          `gorm:"column:id;primaryKey;autoIncrement;not null" json:"id"`                       // bigserial, 主键自动增长
	Address       string          `gorm:"column:address;type:varchar;not null" json:"address"`                         // varchar, 非空
	WtaKey        string          `gorm:"column:wta_key;type:varchar;not null" json:"wtaKey"`                          // varchar, 非空
	ChainID       int8            `gorm:"column:chain_id;type:int" json:"chainID"`                                     // int4, 可以为 nil
	TokenAddress  string          `gorm:"column:token_address;type:varchar" json:"tokenAddress"`                       // varchar, 可以为 nil
	Amount        decimal.Decimal `gorm:"column:amount;type:numeric;default:0;not null" json:"amount"`                 // numeric, 默认值 0, 非空
	Decimals      uint8           `gorm:"column:decimals;type:int" json:"decimals"`                                    // int4, 可以为 nil
	UpdateTime    time.Time       `gorm:"column:update_time;type:timestamp" json:"updateTime"`                         // timestamp, 可以为 nil
	ChainCode     string          `gorm:"column:chain_code;type:varchar" json:"chainCode"`                             // varchar, 可以为 nil
	InAmount      decimal.Decimal `gorm:"column:in_amount;type:numeric;default:0;not null" json:"inAmount"`            // numeric, 默认值 0, 非空
	OutAmount     decimal.Decimal `gorm:"column:out_amount;type:numeric;default:0;not null" json:"outAmount"`          // numeric, 默认值 0, 非空
	InAmountSwap  decimal.Decimal `gorm:"column:in_amount_swap;type:numeric;default:0;not null" json:"inAmountSwap"`   // numeric, 默认值 0, 非空
	OutAmountSwap decimal.Decimal `gorm:"column:out_amount_swap;type:numeric;default:0;not null" json:"outAmountSwap"` // numeric, 默认值 0, 非空
	InUsdSwap     decimal.Decimal `gorm:"column:in_usd_swap;type:numeric;default:0;not null" json:"inUsdSwap"`         // numeric, 默认值 0, 非空
	OutUsdSwap    decimal.Decimal `gorm:"column:out_usd_swap;type:numeric;default:0;not null" json:"outUsdSwap"`       // numeric, 默认值 0, 非空
	Account       string          `gorm:"-"`
	Up            bool            `gorm:"-"`
}

// 实现 ToString 方法
func (t *TokenHold) ToString() string {

	marshal, err := json.Marshal(t)
	if err != nil {
		return fmt.Sprintf("TokenHold: %s", fmt.Sprintf("%+v", t))

	}
	return fmt.Sprintf("TokenHold: %s", string(marshal))
}

// TransferRecord 表示 transfer_record 表的 GORM 结构体
type TransferRecord struct {
	Time       time.Time `gorm:"column:time;type:timestamp;not null" json:"time"`
	Tx         string    `gorm:"column:tx;type:varchar;not null" json:"tx"`
	From       string    `gorm:"column:from;type:varchar;not null" json:"from"`
	To         string    `gorm:"column:to;type:varchar;not null" json:"to"`
	Value      uint64    `gorm:"column:value;type:numeric;not null;default:0" json:"value"`
	Block      uint64    `gorm:"column:block;type:int4" json:"block"`
	GasUsed    uint64    `gorm:"column:gas_used;type:numeric" json:"gasUsed"`
	GasPrice   uint64    `gorm:"column:gas_price;type:numeric;default:0" json:"gasPrice"`
	IsContract bool      `gorm:"column:is_contract;type:bool" json:"isContract"`
	TxTime     int64     `gorm:"column:tx_time;type:int8" json:"txTime"`
	Contract   string    `gorm:"column:contract;type:varchar" json:"contract"`
	ChainID    int64     `gorm:"column:chain_id;type:int8" json:"chainId"`
	ChainCode  string    `gorm:"column:chain_code;type:varchar" json:"chainCode"`
	Authority  string    `gorm:"column:authority;type:varchar" json:"authority"`
	Decimals   uint8     `gorm:"column:decimals;type:int2" json:"decimals"`
}

// TokenFlow 表示 token_flow 表的 GORM 结构体
type TokenFlow struct {
	Time           time.Time       `gorm:"column:time;type:timestamp;not null" json:"time"`
	Pair           string          `gorm:"column:pair;type:varchar;not null" json:"pair"`
	From           string          `gorm:"column:from;type:varchar;not null" json:"from"`
	To             string          `gorm:"column:to;type:varchar;not null" json:"to"`
	Type           string          `gorm:"column:type;type:varchar;not null" json:"type"`
	Block          uint64          `gorm:"column:block;type:int8" json:"block"`
	Token0         string          `gorm:"column:token0;type:varchar" json:"token0"`
	Token1         string          `gorm:"column:token1;type:varchar" json:"token1"`
	Amount0In      decimal.Decimal `gorm:"column:amount0_in;type:numeric;default:0" json:"amount0In"`
	Amount0Out     decimal.Decimal `gorm:"column:amount0_out;type:numeric;default:0" json:"amount0Out"`
	Amount1In      decimal.Decimal `gorm:"column:amount1_in;type:numeric;default:0" json:"amount1In"`
	Amount1Out     decimal.Decimal `gorm:"column:amount1_out;type:numeric;default:0" json:"amount1Out"`
	Tx             string          `gorm:"column:tx;type:varchar" json:"tx"`
	GasPrice       float64         `gorm:"column:gas_price;type:numeric;default:0" json:"gasPrice"`
	GasUse         uint64          `gorm:"column:gas_use;type:numeric;default:0" json:"gasUse"`
	TxTime         int64           `gorm:"column:tx_time;type:int8" json:"txTime"`
	ChainID        int64           `gorm:"column:chain_id;type:int8" json:"chainId"`
	Price          decimal.Decimal `gorm:"column:price;type:numeric;default:0" json:"price"`
	ChainCode      string          `gorm:"column:chain_code;type:varchar" json:"chainCode"`
	UniqueID       int64           `gorm:"column:unique_id;type:bigserial;not null;autoIncrement" json:"uniqueId"`
	Sender         string          `gorm:"column:sender;type:varchar" json:"sender"`
	Recipient      string          `gorm:"column:recipient;type:varchar" json:"recipient"`
	Token0Decimals uint8           `gorm:"-" json:"token0Decimals"`
	Token1Decimals uint8           `gorm:"-" json:"token1Decimals"`
	Retry          uint8           `gorm:"-" json:"retry"`
	Chg            []BalanceChange `gorm:"-" json:"chg"`
}
type Pool struct {
	ID                 int64           `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Token0             string          `gorm:"column:token0;type:varchar" json:"token0"`
	Token1             string          `gorm:"column:token1;type:varchar" json:"token1"`
	Pair               string          `gorm:"column:pair;type:varchar;not null" json:"pair"`
	Dex                string          `gorm:"column:dex;type:varchar" json:"dex"`
	Amount0            decimal.Decimal `gorm:"column:amount0;type:numeric;default:0" json:"amount0"`
	Amount1            decimal.Decimal `gorm:"column:amount1;type:numeric;default:0" json:"amount1"`
	CreateAt           int64           `gorm:"column:create_at;type:bigint" json:"createAt"`
	Fee                int32           `gorm:"column:fee;type:int;default:0" json:"fee"`
	Tvl                decimal.Decimal `gorm:"column:tvl;type:numeric;default:0" json:"tvl"`
	BaseToken          string          `gorm:"column:base_token;type:varchar" json:"baseToken"`
	QuoteToken         string          `gorm:"column:quote_token;type:varchar" json:"quoteToken"`
	ChainCode          string          `gorm:"column:chain_code;type:varchar" json:"chainCode"`
	Authority          string          `gorm:"column:authority;type:varchar" json:"authority"`
	Liq                decimal.Decimal `gorm:"column:liq;type:numeric" json:"liq"`
	StartPrice         decimal.Decimal `gorm:"column:start_price;type:numeric" json:"startPrice"`
	BaseTokenDecimals  uint8           `gorm:"-" json:"baseTokenDecimals"`
	QuoteTokenDecimals uint8           `gorm:"-" json:"quoteTokenDecimals"`
}
type PoolEvent struct {
	ID                  int64           `gorm:"column:id;primaryKey;autoIncrement" `
	Pair                string          `gorm:"column:pair;type:varchar" json:"pair"`
	Token0              string          `gorm:"column:token0;type:varchar" json:"token0"`
	Token1              string          `gorm:"column:token1;type:varchar" json:"token1"`
	Amount0             decimal.Decimal `gorm:"column:amount0;type:numeric;default:0" json:"amount0"`
	Amount1             decimal.Decimal `gorm:"column:amount1;type:numeric;default:0" json:"amount1"`
	Liquidity           decimal.Decimal `gorm:"column:liquidity;type:numeric;default:0" json:"liquidity"`
	AddLiquidityAddress string          `gorm:"column:add_liquidity_address;type:varchar" json:"addLiquidityAddress"`
	Tx                  string          `gorm:"column:tx;type:varchar" json:"tx"`
	Block               uint64          `gorm:"column:block;type:bigint" json:"block"`
	TxTime              int64           `gorm:"column:tx_time;type:bigint" json:"txTime"`
	Type                string          `gorm:"column:type;type:varchar" json:"type"`
	Dex                 string          `gorm:"column:dex;type:varchar" json:"dex"`
	ChainCode           string          `gorm:"column:chain_code;type:varchar" json:"chainCode"`
	CoinAccount         string          `gorm:"-"`
	PcAccount           string          `gorm:"-"`
}

func (e PoolEvent) ToString() string {
	// 实现 ToString 方法

	marshal, err := json.Marshal(e)
	if err != nil {
		//toString方法出错，手动设置toString方法返回值
		return fmt.Sprintf("PoolEvent{ID:%d,PairAddress:%s,Token0:%s,Token1:%s,Amount0:%s,Amount1:%s,Liquidity:%s,AddLiquidityAddress:%s,Tx:%s,Block:%d,TxTime:%d,Type:%s,Dex:%s,ChainCode:%s,CoinAccount:%s,PcAccount:%s}", e.ID, e.Pair, e.Token0, e.Token1, e.Amount0.String(), e.Amount1.String(), e.Liquidity.String(), e.AddLiquidityAddress, e.Tx, e.Block, e.TxTime, e.Type, e.Dex, e.ChainCode, e.CoinAccount, e.PcAccount)
	}
	return string(marshal)
}

type TvKlineData struct {
	Time      time.Time       `gorm:"column:time;not null"`                      // 时间戳，不允许为空
	Pair      string          `gorm:"column:pair;not null"`                      // 交易对，不允许为空
	ChainCode string          `gorm:"column:chain_code;not null"`                // 链代码，不允许为空
	Price     decimal.Decimal `gorm:"column:price;default:0"`                    // 价格，默认值为0
	Amount    decimal.Decimal `gorm:"column:amount"`                             // 交易量，可以为空
	Buy       bool            `gorm:"column:buy"`                                // 购买标志，可以为空（使用指针以支持NULL）
	UniqueID  uint64          `gorm:"column:unique_id;primaryKey;autoIncrement"` // 唯一标识，主键，自动增长
}

func (PoolEvent) TableName() string {
	return "sol_pool_event"
}

func (TvKlineData) TableName() string {
	return "sol_tv_kline_data"
}

// TableName 指定表名为 sol_pools
func (Pool) TableName() string {
	return "sol_pools"
} // TableName 指定表名为 token_flow
func (TokenFlow) TableName() string {
	return "sol_token_flow"
}
func (TokenHold) TableName() string {
	return "sol_token_holds"
}

// TableName 指定表名为 transfer_record
func (TransferRecord) TableName() string {
	return "sol_transfer_record"
}
func (Token) TableName() string {
	return "sol_tokens"
}

type BalanceChange struct {
	Owner    string          `json:"owner"`    // 用户钱包
	Token    string          `json:"token"`    //代币地址
	Pre      decimal.Decimal `json:"pre"`      //交易前余额
	Post     decimal.Decimal `json:"post"`     //交易后余额
	Chg      decimal.Decimal `json:"chg"`      //余额变化量 绝对值
	Decimals int8            `json:"decimals"` // 代币精度
}
type OkxSwapInfo struct {
	Sender            string          `json:"sender"`
	TokenFrom         string          `json:"tokenFrom"`
	TokenTo           string          `json:"tokenTo"`
	TokenFromDecimals uint8           `json:"tokenFromDecimals"`
	TokenToDecimals   uint8           `json:"tokenToDecimals"`
	AmountFrom        decimal.Decimal `json:"amountFrom"`
	AmountTo          decimal.Decimal `json:"amountTo"`
	Price             decimal.Decimal `json:"price"`
	Tx                string          `json:"tx"`
	Block             uint64          `json:"block"`
	BlockTime         int64           `json:"blockTime"`
	GasUsed           uint64          `json:"gasUsed"`
	PubTime           int64           `json:"pubTime"`
	PubTimeStr        string          `json:"pubTimeStr"`
	TokenFromChg      BalanceChange   `json:"tokenFromChg"`
	TokenToChg        BalanceChange   `json:"tokenToChg"`
}
