package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/powershitxyz/SolanaProbe/database"
	"github.com/powershitxyz/SolanaProbe/sys"

	"github.com/shopspring/decimal"
)

var httpClient = http.Client{Timeout: 10 * time.Second}
var QuoteMap = make(map[string]*QuoteToken)

// todo: 自定义的报价币(非常用稳定币)  添加到CustomQuote  以增加对山寨币:山寨币的价格获取
// todo:  被动触发山寨币:山寨币的价格获取时加载 拿tvl最大的池子的价格 并定时更新最大池子地址?多久更新一次?
// todo: swap更新池子价格
var CustomQuote = make(map[string]*QuoteToken)
var quoteKey = "quote"

func init() {
	var devFlag = true
	if configFilePathFromEnv := os.Getenv("DALINK_GO_CONFIG_PATH"); configFilePathFromEnv != "" {
		devFlag = false
	} else {
		devFlag = true
	}

	QuoteMap["So11111111111111111111111111111111111111112"] = &QuoteToken{
		Address:     "So11111111111111111111111111111111111111112",
		PairAddress: "Czfq3xZZDmsdGdUyrNLtRhGc47cXcZtLG4crryfu44zE",
		Symbol:      "SOL",
		ChainCode:   "SOLANA",
		Decimals:    9,
		PairSymbol:  "SOLUSDT",
		Price:       decimal.NewFromFloat(0),
		LestTime:    0,
		Mainnet:     true,
		Logo:        "https://img.apihellodex.lol/BSC/0x570a5d26f7765ecb712c0924e4de545b89fd43df.png",
	}
	QuoteMap["EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"] = &QuoteToken{
		Address:    "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
		Symbol:     "USDC",
		PairSymbol: "USDC",
		Decimals:   6,
		Price:      decimal.NewFromFloat(1),
	}
	QuoteMap["Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB"] = &QuoteToken{
		Address:    "Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB",
		Symbol:     "USDT",
		PairSymbol: "USDC",
		Decimals:   6,
		Price:      decimal.NewFromFloat(1),
	}
	//获取除去稳定币其他币种的价格
	reqQuoteMap := make(map[string][]*QuoteToken)
	for _, token := range QuoteMap {
		if token.PairSymbol != "USDC" {
			tokens, exies := reqQuoteMap[token.PairSymbol]
			if exies {
				tokens = append(tokens, token)
				reqQuoteMap[token.PairSymbol] = tokens
			}
			reqQuoteMap[token.PairSymbol] = []*QuoteToken{token}
		}

	}
	var reqQuoteSlice = make([]string, 0)
	for key := range reqQuoteMap {
		reqQuoteSlice = append(reqQuoteSlice, key)
	}
	//fmt.Println("Error marshaling JSON: ", reqQuoteSlice)
	jsonData, err := json.Marshal(reqQuoteSlice)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	fmt.Println("req params:", string(jsonData))

	// URL encode the JSON string
	encodedData := url.QueryEscape(string(jsonData))
	fmt.Println("req paramsEncoded:", fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbols=%s", encodedData))
	client := &http.Client{}
	if devFlag {
		proxyURL, err := url.Parse("http://127.0.0.1:7890")
		if err != nil {
			panic(err)
		}
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
		client = &http.Client{
			Transport: transport,
		}
	}

	resp, err := client.Get(fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbols=%s", encodedData))
	if err != nil {
		fmt.Println("req error:", err)
	}
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic: %v", r)
		}
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("req error:", err)
	}
	if err != nil {
		fmt.Println("resp error: data:", string(jsonData))
	}
	var ps = []PairSymbolPrice{}
	fmt.Println("resp body :", string(body))
	if err := json.Unmarshal(body, &ps); err != nil {
		sys.Logger.Error("resp error:", err, "data:", string(body))

		res := &QuoteToken{}
		database.HGet(quoteKey, "SOLANA:So11111111111111111111111111111111111111112", res)
		sys.Logger.Infof("redisData %s---:%+v:", time.Unix(0, res.LestTime*int64(time.Millisecond)).Format("2006-01-02 15:04:05.000"), res)
		ps = append(ps, PairSymbolPrice{
			Symbol: "SOLUSDT",
			Price:  res.Price,
		})
	}
	if ps[0].Price.LessThanOrEqual(decimal.NewFromFloat(0)) {
		panic("Sol price less than 0")
	}
	for _, p := range ps {
		tokens, has := reqQuoteMap[p.Symbol]
		if has {
			for _, token := range tokens {
				if token.PairSymbol == p.Symbol {
					token.SetPrice(p.Price)
					//quoteToken := QuoteMap[token.Address]
					//quoteToken.SetPrice(p.Price)
				}
			}
		}
		// todo 报价币入缓存!!!!!!!!!
		fmt.Println("resp ", p)
	}
	// 序列化为 JSON
	jsonDataAll, err := json.Marshal(QuoteMap)
	if err != nil {
		log.Fatalf("JSON 序列化失败: %s", err)
	}
	fmt.Println("quoteMap: ", string(jsonDataAll))
}

type QuoteToken struct {
	Address     string          `json:"address"`
	PairAddress string          `json:"pairAddress"`
	ChainCode   string          `json:"chainCode"`
	Symbol      string          `json:"symbol"`
	Decimals    int8            `json:"decimals"`
	PairSymbol  string          `json:"pairSymbol"`
	Price       decimal.Decimal `json:"price"`
	LestTime    int64           `json:"lastTime"`
	Mainnet     bool            `json:"mainnet"`
	Logo        string          `json:"logo"`
}

// 实现 String 方法
func (q QuoteToken) String() string {
	return fmt.Sprintf(
		"Address: %s, PairAddress: %s, ChainCode: %s, Symbol: %s, Decimals: %d, PairSymbol: %s, Price: %s, LestTime: %d, Mainnet: %t",
		q.Address, q.PairAddress, q.ChainCode, q.Symbol, q.Decimals, q.PairSymbol, q.Price.String(), q.LestTime, q.Mainnet,
	)
}

type PairSymbolPrice struct {
	Symbol string          `json:"symbol"`
	Price  decimal.Decimal `json:"price"`
}

// SetPrice sets the price and updates the lestTime to the current timestamp.
func (qt *QuoteToken) SetPrice(price decimal.Decimal) {
	defer func() {
		if r := recover(); r != nil {
			Log.Errorf("quoteMapError: %s,%s  %v", qt.String(), price.String(), r)
		}
	}()
	qt.Price = price
	qt.LestTime = time.Now().UnixMilli()
	if qt.PairSymbol != "USDC" {
		updateQuote(*qt)
	}

}

func GetAll() map[string]*QuoteToken {
	return QuoteMap
}

func updateQuote(q QuoteToken) {
	database.HSet(quoteKey, q.ChainCode+":"+q.Address, q)
}
