package pub

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/powershitxyz/SolanaProbe/config"
	"github.com/powershitxyz/SolanaProbe/model"
	"github.com/powershitxyz/SolanaProbe/sys"

	"strings"
	"time"

	"github.com/shopspring/decimal"
)

var Conf = config.GetConfig()
var RedisExpireTimeByDay time.Duration = time.Hour * 24 * 8 // 30 days

var MaxSlot uint64 = 0
var WSOL = "So11111111111111111111111111111111111111112"
var WSOLToken = &model.Token{
	ID:          600673,
	Name:        "Wrapped SOL",
	Symbol:      "SOL",
	Address:     WSOL,
	Decimals:    9,
	TotalSupply: 0,
	ChainCode:   "SOLANA",
}

var SOL = "11111111111111111111111111111111"
var USDC = "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
var USDT = "Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB"
var SOLStr = "SOL"
var Log = sys.Logger
var Transfer = "Transfer"
var SOLANA = "SOLANA"
var InitializeAccount = "InitializeAccount"
var InitializeAccount2 = "InitializeAccount2"
var InitializeAccount3 = "InitializeAccount3"
var SysTransfer = "STransfer"
var TransferChecked = "TransferChecked"
var BurnChecked = "BurnChecked"

// 事件标记 标记
var LiqRemove = "LiqRemove"
var LiqAdd = "LiqAdd"
var LiqCreate = "LiqCreate"
var LiqSwap = "LiqSwap"

// dex
var DexRaydium = "Raydium"
var DexOrca = "Orca"
var DexOther = "Other"
var DexOkxProxy = "OkxProxy"
var DexJupiterV6 = "JupiterV6"
var DexMeteora = "Meteora"
var DexFluxbeam = "Fluxbeam"
var DexPumpFun = "PumpFun"
var DexMoonshot = "Moonshot"

// event
var LIQ_ADD = "MINT"
var LIQ_REMOVE = "BURN"
var LIQ_CREATE = "NEW_POOL"
var VotePrefix = "Program Vote1111"
var Vote = "Vote111111111111111111111111111111111111111"

// 手续费地址

// pumbfun迁移账号
var PumpfunRaydiumMigration = "39azUYFWPz3VHgKCf3VChUwbpURdCHRxjWVowf5jUJjg"

// topic
var TokenFlowTopic = "t:h_flow"
var TransferTopic = "t:h_s_transfer"
var PoolEventTopic = "t:h_s_poolevent"
var PoolCreateTopic = "t:h_s_poolCreate"
var PoolUpdateTopic = "t:h_s_poolUpdate"
var NcTokenMetaTopic = "token.meta"
var TokenMetaTopic = "t:token_meta"

// okx 交易信息推送
var OkxTradeTopic = "t:okx_trade"

type ParsedData struct {
	Transfers      []model.TransferRecord
	TokenFlow      []model.TokenFlow
	TokenHold      []model.TokenHold
	HoldsUpdateKey map[int][]string
	NewPairs       []Pair
	UpdatedPairs   []Pair
	PoolEvents     []model.PoolEvent
}

type DecodeInsTempData struct {
	TxSlot           uint64
	TxTime           time.Time
	Tx               string
	TempAccount      map[string]TempAccountData
	TokenDecimalsMap map[string]uint8
	Fee              uint64
	TokenHolds       []model.TokenHold
	SOLHolds         map[string]uint64
}
type TempAccountData struct {
	Account  string
	Mint     string
	Owner    string
	Decimals uint8
	TxTime   time.Time
	Block    uint64
}

type SwapInfo struct {
	From               string
	Pair               string
	Dex                string
	In                 string
	Out                string
	CoinAccount        string
	PcAccount          string
	CoinAccountTrue    string
	PcAccountTrue      string
	AmountIn           decimal.Decimal
	AmountOut          decimal.Decimal
	CoinAmount         decimal.Decimal
	QuoteAmount        decimal.Decimal
	PriceMutQuotePrice decimal.Decimal
	Price              decimal.Decimal
	BaseToken          string
	QuoteToken         string
	BaseDecimals       uint8
	QuoteDecimals      uint8
	TxTime             int64
	TxHash             string
	Slot               uint64
	Curr               int64
}

// 实现 ToString 方法
func (r *SwapInfo) ToString() string {

	marshal, err := json.Marshal(r)
	if err != nil {
		return fmt.Sprintf("SwapInfo: %s", fmt.Sprintf("%+v", r))

	}
	return fmt.Sprintf("SwapInfo: %s", string(marshal))
}

type Pair struct {
	Id           int64
	Pair         string
	Authority    string
	Dex          string
	BaseToken    string
	QuoteToken   string
	BaseDecimal  uint8
	QuoteDecimal uint8
	CoinAmount   decimal.Decimal
	QuoteAmount  decimal.Decimal
	Liq          decimal.Decimal
	Tvl          decimal.Decimal
	InDb         bool
	LastUpdate   int64
	Count        int64
	CreatedAt    int64
	StartPrice   decimal.Decimal
}

func (qt *Pair) SetCount(setZero bool) {
	if setZero {
		qt.Count = 0
	}
	qt.Count = qt.Count + 1
}
func (qt *Pair) SetAmount(amountBase decimal.Decimal, amountQuote decimal.Decimal) {
	qt.QuoteAmount = amountQuote
	qt.CoinAmount = amountBase
	qt.LastUpdate = time.Now().Unix()
}
func (qt *Pair) SetId(id int64) {
	qt.Id = id
}
func (qt *Pair) SetTvl(tvl decimal.Decimal) {
	qt.Tvl = tvl
	qt.LastUpdate = time.Now().Unix()
}

// TokenSearch represents the structure for token search.
type TokenSearch struct {
	ChainId       int       `json:"chainId"`       // 链的代码
	Chain         string    `json:"chain"`         // 链的代码
	ChainCode     string    `json:"chainCode"`     // 链的代码
	BaseAddress   string    `json:"baseAddress"`   // 基础地址
	QuoteAddress  string    `json:"quoteAddress"`  // 报价地址
	BaseDecimals  int       `json:"baseDecimals"`  // 基础代币小数位
	QuoteDecimals int       `json:"quoteDecimals"` // 报价代币小数位
	PairAddress   string    `json:"pairAddress"`   // 对地址
	BaseToken     string    `json:"baseToken"`     // 基础代币
	QuoteToken    string    `json:"quoteToken"`    // 报价代币
	Details       string    `json:"details"`       // poolKey
	AddTime       time.Time `json:"addTime"`       // 添加时间
}
type Period string

const (
	M1  Period = "1m"
	M5  Period = "5m"
	M15 Period = "15m"
	M30 Period = "30m"
	H1  Period = "1h"
	H4  Period = "4h"
	D1  Period = "1d"
	W1  Period = "1w"
)

var Periods = []Period{M1, M5, M15, M30, H1, H4, D1, W1}

type KChart struct {
	O         string          `json:"O"`         // 开盘价
	C         string          `json:"C"`         // 收盘价
	H         string          `json:"H"`         // 最高价
	L         string          `json:"L"`         // 最低价
	Timestamp int64           `json:"timestamp"` // 时间戳
	Time      string          `json:"time"`      // 可读时间
	Volume    decimal.Decimal `json:"volume"`    // 交易量
	VolumeUsd decimal.Decimal `json:"volumeUsd"` // 交易额美元
}

func GetPKFromBase58(str string) *solana.PublicKey {
	base58 := solana.MustPublicKeyFromBase58(str)
	return &base58
}

// reverseBytes 翻转字节数组
func ReverseBytes(data []byte) []byte {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
	return data
}

const emptyHash = "000000"

func HashKey(str1, str2 string) string {
	if len(str1) < 5 {
		str1 = emptyHash
	}
	if len(str2) < 3 {
		str1 = emptyHash
	}
	// 获取前 4 个字符并转换为小写
	prefix1 := strings.ToLower(str1[:5])
	prefix2 := strings.ToLower(str2[:3])

	// 计算 SHA-256 的哈希值
	hash := sha256.Sum256([]byte(str1 + str2))
	hashHex := hex.EncodeToString(hash[:])

	// 提取哈希值的前 26 个字符
	hashPrefix := hashHex[:24]

	// 拼接结果
	return prefix1 + prefix2 + hashPrefix
}

// 定义一个通用的分批函数
func BatchSlice[T any](s []T, batchSize int) [][]T {
	if batchSize <= 0 {
		return [][]T{} // 如果批大小不合法，返回 nil
	}

	batches := make([][]T, 0)
	for i := 0; i < len(s); i += batchSize {
		end := i + batchSize
		if end > len(s) {
			end = len(s) // 确保不超出切片范围
		}
		batches = append(batches, s[i:end]) // 添加子切片到结果切片
	}
	return batches
}
