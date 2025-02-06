package dego

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/klauspost/compress/gzhttp"
	"github.com/powershitxyz/SolanaProbe/config"
	"github.com/powershitxyz/SolanaProbe/pub"
	"github.com/powershitxyz/SolanaProbe/sys"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	addressLookupTable "github.com/gagliardetto/solana-go/programs/address-lookup-table"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/jsonrpc"
)

type HelloRpc struct {
	rpc     *rpc.Client
	url     string
	quote   int
	healthy bool
}

// todo 初始化n个rpc客户端   请求使用的rpc客户端应为轮训rpc客户端列表
var rpcEndpoints = config.GetConfig().Chain.GetRpcMapper()
var rpcUsage sync.Map
var conf = config.GetConfig()

// 全局变量
var (
	rpcClients []HelloRpc
	// clientIndex     int
	mu              sync.Mutex
	usageCh         chan string
	maxConsumers    int32 = 50
	activeConsumers int32 = 0
	rpcUsageMu      sync.Mutex

	otherClient *HelloRpc
)

const (
	lhcThresh = 1 * 60 * 60
)

func init() {
	initClients()
	usageCh = make(chan string, 100)
	go clearUsageEverySecond()
	for i := 0; i < int(maxConsumers); i++ {
		go processUsage()
		atomic.AddInt32(&activeConsumers, 1)
	}
}

func ReadClientMsg() map[string]int {
	result := make(map[string]int)
	rpcUsageMu.Lock()
	rpcUsage.Range(func(key, value interface{}) bool {
		result[key.(string)] = value.(int)
		return true
	})
	rpcUsageMu.Unlock()
	return result
}

func initClients() {
	for _, url := range rpcEndpoints {
		client := rpc.NewWithCustomRPCClient(jsonrpc.NewClientWithOpts(url.Rpc, &rpcClientOpts))
		rpcClients = append(rpcClients, HelloRpc{
			rpc:     client,
			url:     url.Rpc,
			quote:   url.Quote,
			healthy: true,
		})
	}
	initOtherClient()
	go removeUnhealthyClient()
}

func initOtherClient() {
	var rpcStr = "https://rpc.ankr.com/solana/93b55d4324fed365b92a4f7735a790b8d5f3b4962298d41a58cb03b5fab7d847"
	client := rpc.NewWithCustomRPCClient(jsonrpc.NewClientWithOpts(rpcStr, &rpcClientOpts))

	otherClient = &HelloRpc{
		rpc:     client,
		url:     rpcStr,
		quote:   50,
		healthy: true,
	}
}

// todo 轮训rpc客户端列表，移除不健康的客户端
//
//	func removeUnhealthyClient() {
//		for i, client := range rpcClients {
//			time.Sleep(time.Minute * 60)
//			mu.Lock()
//			_, err := client.rpc.GetHealth(context.TODO())
//			if err != nil {
//				client.rpc.Close()
//				rpcClients = append(rpcClients[:i], rpcClients[i+1:]...)
//			}
//			mu.Unlock()
//		}
//	}
func removeUnhealthyClient() {
	for {
		time.Sleep(time.Minute * 60)
		mu.Lock()
		if len(rpcClients) <= 1 {
			mu.Unlock()
			return
		}
		for i := len(rpcClients) - 1; i >= 0; i-- {
			client := rpcClients[i]
			_, err := client.rpc.GetHealth(context.TODO())
			if err != nil {
				// rpcClients = append(rpcClients[:i], rpcClients[i+1:]...)
				rpcClients[i].healthy = false
				sys.Logger.Info("Removed and Marked unhealthy client:", client.url, err)
			} else {
				if !rpcClients[i].healthy {
					rpcClients[i].healthy = true
				}
			}
		}
		mu.Unlock()
	}
}

func selectClient() (*HelloRpc, error) {
	for _, v := range rpcClients {
		if v.healthy {
			return &v, nil
		}
	}
	return nil, errors.New("no available")
}

func Client() *rpc.Client {
	for {
		mu.Lock()
		// client := rpcClients[clientIndex]
		client, err := selectClient()
		if err != nil {
			mu.Unlock()
			return nil
		}
		rpcUsageMu.Lock()
		quoteVal, ok := rpcUsage.Load(client.url)
		rpcUsageMu.Unlock()
		if !ok {
			quoteVal = 0
		}

		if client.quote == 0 || quoteVal.(int) < client.quote {
			select {
			case usageCh <- client.url:
				_ = client.url
			default:
				// log.Printf("usageCh is full, skipping update for %s", client.url)
				addConsumer()
			}
			mu.Unlock()
			return client.rpc
		} else {
			sys.Logger.Errorf("rpc: %s over quote: %d", client.url, quoteVal.(int))
			// clientIndex = (clientIndex + 1) % len(rpcClients) // 只有在quote不足时才更新clientIndex
		}
		mu.Unlock()
	}
}

func addConsumer() {
	if atomic.LoadInt32(&activeConsumers) < maxConsumers {
		atomic.AddInt32(&activeConsumers, 1)
		go processUsage()
		log.Printf("Added a new consumer, total: %d", atomic.LoadInt32(&activeConsumers))
	}
}

func processUsage() {
	defer atomic.AddInt32(&activeConsumers, -1)
	for url := range usageCh {
		rpcUsageMu.Lock()
		currentVal, _ := rpcUsage.LoadOrStore(url, 0)
		rpcUsage.Store(url, currentVal.(int)+1)
		rpcUsageMu.Unlock()
	}
}

func clearUsageEverySecond() {
	for range time.Tick(time.Second) {
		rpcUsageMu.Lock()
		rpcUsage.Range(func(key, value interface{}) bool {
			rpcUsage.Store(key, 0)
			return true
		})
		rpcUsageMu.Unlock()
	}
}

var supplyLimit = uint64(0)

func GetTokenSupply(ctx context.Context, account solana.PublicKey, status rpc.CommitmentType) (*rpc.GetTokenSupplyResult, error) {
	start := time.Now().UnixMilli()
	defer func() {
		supplyLimit = supplyLimit + 1
		if supplyLimit%10 == 0 || time.Now().UnixMilli()-start > 1000 {
			if supplyLimit > ^uint64(0)-10000 {
				supplyLimit = 1
			}
			sys.Logger.Infof("耗时1 GetTokenSupply:%dms,参数: %s", time.Now().UnixMilli()-start, account.String())
		}
		if r := recover(); r != nil {
			sys.Logger.Errorf("RPC GetTokenSupply Issue Recovered from panic: %v", r)
		}
	}()

	out, err := Client().GetTokenSupply(ctx, account, status)
	return out, err
}

var accountInfoWithOptsLimit = uint64(0)

func GetAccountInfoWithOpts(ctx context.Context, account solana.PublicKey, opts *rpc.GetAccountInfoOpts) (*rpc.GetAccountInfoResult, error) {
	start := time.Now().UnixMilli()
	defer func() {
		accountInfoWithOptsLimit = accountInfoWithOptsLimit + 1
		if accountInfoWithOptsLimit%20 == 0 || time.Now().UnixMilli()-start > 1000 {
			if accountInfoWithOptsLimit > ^uint64(0)-10000 {
				accountInfoWithOptsLimit = 1
			}
			sys.Logger.Infof("耗时1 GetAccountInfoWithOpts:%dms,目前用在GetTokenMetaInfo参数: %s", time.Now().UnixMilli()-start, account.String())
		}
		if r := recover(); r != nil {
			sys.Logger.Errorf("RPC GetAccountInfoWithOpts Issue Recovered from panic: %v", r)
		}
	}()
	return Client().GetAccountInfoWithOpts(ctx, account, opts)
}

var getAccountInfoLimit = uint64(0)

func GetAccountInfo(ctx context.Context, account solana.PublicKey) (out *rpc.GetAccountInfoResult, err error) {
	start := time.Now().UnixMilli()
	defer func() {
		getAccountInfoLimit = getAccountInfoLimit + 1
		if getAccountInfoLimit%20 == 0 || time.Now().UnixMilli()-start > 1000 {
			if getAccountInfoLimit > ^uint64(0)-10000 {
				getAccountInfoLimit = 1
			}
			sys.Logger.Infof("耗时1 getAccountInfoLimit:%dms,参数: %s", time.Now().UnixMilli()-start, account.String())
		}
		if r := recover(); r != nil {
			sys.Logger.Errorf("RPC GetAccountInfo Issue Recovered from panic: %v", r)
		}
	}()
	return Client().GetAccountInfo(ctx, account)
}

func GetTransactionsBySlot(slot uint64) (decodedTransactions []*RawTransaction, blockWithNoTxs *rpc.GetBlockResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			sys.Logger.Error("GetTransactionsBySlotMightBe229:", r, " unsolved slot:", slot)
		}
	}()
	blockWithNoTxs, err = getBlockWithRetries(slot, 5, 5*time.Second)
	if err != nil {
		sys.Logger.Errorln("GetTransactionsBySlot-Error", slot, err, " especially conds.")
		return nil, nil, err
	}
	for _, transaction := range blockWithNoTxs.Transactions {
		if transaction.Meta.Err == nil {
			decodedTx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(transaction.Transaction.GetBinary()))
			if err == nil {
				if len(transaction.Meta.LogMessages[0]) > 1 && strings.HasPrefix(transaction.Meta.LogMessages[0], pub.VotePrefix) && strings.HasPrefix(transaction.Meta.LogMessages[1], pub.VotePrefix) && len(decodedTx.Message.Instructions) == 1 {
					continue
				}
				decodedTransactions = append(decodedTransactions, &RawTransaction{
					Transaction: decodedTx,
					Meta:        transaction.Meta,
					Slot:        slot,
				})
			}
		}

	}
	blockWithNoTxs.Transactions = nil
	return decodedTransactions, blockWithNoTxs, nil
}

var LookupTableLimit = uint64(0)

func GetAddressLookupTableWithRetry(
	ctx context.Context,
	address solana.PublicKey,
) (*addressLookupTable.AddressLookupTableState, error) {
	start := time.Now().UnixMilli()
	RetryCount := 0
	defer func() {
		LookupTableLimit = LookupTableLimit + 1
		if LookupTableLimit%200 == 0 || time.Now().UnixMilli()-start > 5000 {
			if LookupTableLimit > ^uint64(0)-10000 {
				LookupTableLimit = 1
			}
			sys.Logger.Infof("耗时1 GetAddressLookupTableWithRetry:%dms,次数:%d, 参数: %s", time.Now().UnixMilli()-start, RetryCount, address.String())
		}
	}()
	const maxRetries = 3
	var account *rpc.GetAccountInfoResult
	var err error
	for i := 0; i < maxRetries; i++ {
		RetryCount++
		account, err = Client().GetAccountInfo(ctx, address)
		if err == nil && account != nil {
			break
		}
		if i < maxRetries-1 {
			time.Sleep(time.Second * time.Duration(i+1)) // Exponential backoff
		}
	}
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, fmt.Errorf("account not found")
	}
	return addressLookupTable.DecodeAddressLookupTableState(account.GetBinary())
}

var maxVersion = uint64(0)
var (
	defaultMaxIdleConns        = 3000
	defaultMaxIdleConnsPerHost = 3000
	defaultTimeout             = 5 * time.Minute
	defaultKeepAlive           = 5 * time.Minute
)

var rpcClientOpts = jsonrpc.RPCClientOpts{
	HTTPClient: &http.Client{
		Timeout:   defaultTimeout,
		Transport: gzhttp.Transport(newHTTPTransport()),
		//Transport: &CurlTransport{
		//	Transport: http.DefaultTransport,
		//},
	},
}

func newHTTPTransport() *http.Transport {
	return &http.Transport{
		IdleConnTimeout:     defaultTimeout,
		MaxConnsPerHost:     defaultMaxIdleConnsPerHost,
		MaxIdleConnsPerHost: defaultMaxIdleConnsPerHost,
		MaxIdleConns:        defaultMaxIdleConns,
		Proxy:               http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   300 * time.Second,
			KeepAlive: defaultKeepAlive,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		TLSHandshakeTimeout:   20 * time.Second,
		ExpectContinueTimeout: 10 * time.Second,
	}
}

/**************************************************/
type CurlTransport struct {
	Transport http.RoundTripper
}

// RoundTrip executes a single HTTP transaction and logs it
func (t *CurlTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	curlCommand, err := ToCurl(req)
	if err != nil {
		return nil, err
	}
	if strings.Contains(curlCommand, "getBlock") {
		fmt.Println("Generated curl command:", curlCommand)
	}

	// 如果 Parent 为 nil，使用默认的 Transport
	transport := t.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewBuffer(body))

	go saveResponseBody(body, curlCommand)
	//fmt.Printf("Response body saved to: %s\n", filename)

	return resp, nil
}

// ToCurl converts an HTTP request to a curl command string
func ToCurl(req *http.Request) (string, error) {
	curlCommand := "curl -X " + req.Method

	// Add headers
	for key, values := range req.Header {
		for _, value := range values {
			curlCommand += fmt.Sprintf(" -H '%s: %s'", key, value)
		}
	}

	// Add body if present
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return "", err
		}
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Reset the body after reading
		if len(bodyBytes) > 0 {
			curlCommand += fmt.Sprintf(" -d '%s'", string(bodyBytes))
		}
	}

	// Add URL
	curlCommand += fmt.Sprintf(" '%s'", req.URL.String())

	return curlCommand, nil
}

func saveResponseBody(body []byte, curlCmd string) (string, error) {
	if err := os.MkdirAll(ResponseDir, 0755); err != nil {
		return "", fmt.Errorf("error creating directory: %w", err)
	}

	filename := generateFilename(curlCmd)
	if !strings.Contains(filename, "getBlock") {
		return "", nil // 不保存文件，但也不返回错误
	}

	if err := os.MkdirAll(ResponseDir, 0755); err != nil {
		return "", fmt.Errorf("error creating directory: %w", err)
	}

	fullPath := filepath.Join(ResponseDir, filename)
	if err := os.WriteFile(fullPath, body, 0644); err != nil {
		return "", fmt.Errorf("error writing file: %w", err)
	}

	return fullPath, nil
}

var ResponseDir = "/app/hellosol/responses"

func generateFilename(curlCmd string) string {
	dataRegex := regexp.MustCompile(`-d '(.*?)'`)
	matches := dataRegex.FindStringSubmatch(curlCmd)
	if len(matches) < 2 {
		return "unknown_request.txt"
	}

	jsonData := matches[1]

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return "invalid_json_request.txt"
	}

	method, ok := data["method"].(string)
	if !ok {
		method = "unknown_method"
	}

	params, ok := data["params"].([]interface{})
	var paramNumber string
	if ok && len(params) > 0 {
		if num, ok := params[0].(float64); ok {
			paramNumber = fmt.Sprintf("%.0f", num)
		}
	}
	if paramNumber == "" {
		paramNumber = "unknown_param"
	}

	timestamp := time.Now().Format("20060102_150405")
	return fmt.Sprintf("%s_%s_%s.json", method, paramNumber, timestamp)
}
func getBlockWithRetries(slot uint64, maxRetries int, retryDelay time.Duration) (*rpc.GetBlockResult, error) {
	includeRewards := false
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)
	randomDuration := time.Duration(random.Intn(conf.Chain.GetTxDelay())+1000) * time.Millisecond
	time.Sleep(randomDuration)

	var blockWithNoTxs *rpc.GetBlockResult
	var err error

	for retries := 0; retries < maxRetries; retries++ {
		blockWithNoTxs, err = Client().GetBlockWithOpts(
			context.TODO(),
			slot,
			&rpc.GetBlockOpts{
				Encoding:                       solana.EncodingBase64,
				Commitment:                     rpc.CommitmentFinalized,
				TransactionDetails:             rpc.TransactionDetailsFull,
				MaxSupportedTransactionVersion: &maxVersion,
				Rewards:                        &includeRewards,
			},
		)

		if err == nil {
			// Block retrieved successfully
			return blockWithNoTxs, nil
		}

		// Check if the error is due to the block not being available
		rpcErr, ok := err.(*jsonrpc.RPCError)
		if ok {
			if rpcErr.Code == -32004 {
				log.Printf("[ImportantBlockInfo]: %d retry Block not available, retrying in %v...", slot, retryDelay)
			} else {
				log.Printf("[ImportantBlockInfo with other errorcode]: %d retry Block not available, retrying in %v... %v", slot, retryDelay, rpcErr)
			}
		}
		time.Sleep(retryDelay)
		continue
		// If the error is not related to block availability, return it
		//return nil, err
	}

	if otherClient != nil {
		blockWithNoTxs, err = otherClient.rpc.GetBlockWithOpts(
			context.TODO(),
			slot,
			&rpc.GetBlockOpts{
				Encoding:                       solana.EncodingBase64,
				Commitment:                     rpc.CommitmentFinalized,
				TransactionDetails:             rpc.TransactionDetailsFull,
				MaxSupportedTransactionVersion: &maxVersion,
				Rewards:                        &includeRewards,
			},
		)
		if err == nil {
			return blockWithNoTxs, nil
		}
		log.Printf("[ImportantBlockInfo with ankr failed]: %d retry Block not available by call [ANKR] %v...", slot, err)
	}

	return nil, fmt.Errorf("[ImportantBlockInfo-Complete]: %d failed to retrieve block after %d retries", slot, maxRetries)
}
