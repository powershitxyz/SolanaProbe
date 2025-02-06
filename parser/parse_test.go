package parser

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/blocto/solana-go-sdk/common"
	"github.com/blocto/solana-go-sdk/program/token"
	"github.com/blocto/solana-go-sdk/types"
	"github.com/davecgh/go-spew/spew"
	"github.com/mr-tron/base58"
	"github.com/powershitxyz/SolanaProbe/dego"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	addressLookupTable "github.com/gagliardetto/solana-go/programs/address-lookup-table"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/jsonrpc"
	"github.com/klauspost/compress/gzhttp"
)

const rpcUrl = "https://mainnet.helius-rpc.com/?api-key=50465b2c-93d8-4d53-8987-a9ccd7962504"

var (
	defaultMaxIdleConnsPerHost = 50
	defaultTimeout             = 5 * time.Minute
	defaultKeepAlive           = 180 * time.Second
)

func newHTTPTransport() *http.Transport {
	return &http.Transport{
		IdleConnTimeout:     defaultTimeout,
		MaxConnsPerHost:     defaultMaxIdleConnsPerHost,
		MaxIdleConnsPerHost: defaultMaxIdleConnsPerHost,
		Proxy:               http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   defaultTimeout,
			KeepAlive: defaultKeepAlive,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2: true,
		// MaxIdleConns:          100,
		TLSHandshakeTimeout: 10 * time.Second,
		// ExpectContinueTimeout: 1 * time.Second,
	}
}

var opts = jsonrpc.RPCClientOpts{
	HTTPClient: &http.Client{
		Timeout:   defaultTimeout,
		Transport: gzhttp.Transport(newHTTPTransport()),
	},
}
var clientBlock = rpc.NewWithCustomRPCClient(jsonrpc.NewClientWithOpts(rpcUrl, &opts))

func TestParseTx(t *testing.T) {

	maxVersion := uint64(0)
	// txSig := solana.MustSignatureFromBase58("Huwva9qzCP8C8B2gt3DLCGfS15hELxguegeTwoZwfbC2dBruL74aWJvQQuWtpBWR7v2ka846G8sUjReKXrPYRL6") //swap
	txSig := solana.MustSignatureFromBase58("5wum13hgUvG4qnLhoUPPQun9VjaK1igmBcKrhGEWAbt1E6A31Z2fVNMZ8QV9GZH3TWimNA8RVt4XWoLDsE51eNj1") //transfer sol
	// txSig := solana.MustSignatureFromBase58("eWoZtxAQko1APcs71Kcodr4eKxr448kD3TarLqVoMkbTK3PCxN3eq4sp5sFWH514xGxKpVhd4YuRBTVgMBQfEYC") //transfer token

	out, err := clientBlock.GetTransaction(
		context.TODO(),
		txSig,
		&rpc.GetTransactionOpts{
			Encoding:                       solana.EncodingBase64,
			MaxSupportedTransactionVersion: &maxVersion,
		},
	)
	if err != nil {
		panic(err)
	}

	txSlot := out.Slot
	txMeta := out.Meta
	txVer := out.Version
	txTran := out.Transaction

	Log.Println(txSlot, txMeta, txVer, txTran)

	decodedTx, _ := solana.TransactionFromDecoder(bin.NewBinDecoder(txTran.GetBinary()))
	lookupTableAccounts := decodedTx.Message.AddressTableLookups
	for _, lookup := range lookupTableAccounts {
		lookupTable, err := addressLookupTable.GetAddressLookupTable(context.Background(), clientBlock, lookup.AccountKey)
		if err != nil {
			panic(err)
		}

		_ = decodedTx.Message.SetAddressTables(map[solana.PublicKey]solana.PublicKeySlice{
			lookup.AccountKey: lookupTable.Addresses,
		})
	}

	metas, err := decodedTx.Message.AccountMetaList()
	acc := make([]string, 0)
	for _, v := range metas {
		acc = append(acc, v.PublicKey.String())
	}

	for i, _ := range txMeta.InnerInstructions {
		varr := txMeta.InnerInstructions[i]
		for j, _ := range varr.Instructions {
			compiledInstru := varr.Instructions[j]
			buf := &bytes.Buffer{}
			encoder := bin.NewBinEncoder(buf)
			err := encoder.Encode(compiledInstru)
			if err != nil {
				Log.Println(err)
			}

			programPub, _ := decodedTx.Message.ResolveProgramIDIndex(compiledInstru.ProgramIDIndex)

			accounts := make([]*solana.AccountMeta, 0)
			accountsStr := make([]string, 0)
			for _, v := range compiledInstru.Accounts {
				accounts = append(accounts, metas[v])
				accountsStr = append(accountsStr, metas[v].PublicKey.String())
			}
			accountsStr = append(accountsStr, programPub.String())
			result, err := ParseDecode(programPub, accounts, compiledInstru.Data)

			if err != nil {
				Log.Println(err)
			}

			if i == 1 && (j == 2 || j == 1) {
				for ii, _ := range accounts {
					// 获取源账户信息
					sourceAccInfo, err := clientBlock.GetAccountInfo(context.Background(), accounts[ii].PublicKey)
					if err != nil {
						log.Printf("failed to get source account info: %v", err)
					} else {
						// 获取代币合约地址（mint 地址）

						mintAddress := sourceAccInfo.Value.Data.GetBinary()[:32] // mint 地址通常是前 32 个字节
						sma := solana.PublicKeyFromBytes(mintAddress).String()
						log.Printf("Mint Address: %s\n", sma)
					}
				}
			}

			vss := compiledInstru.Data.String()
			spew.Dump(vss, result)
			Log.Println(vss, result)

		}
	}

	for i, _ := range decodedTx.Message.Instructions {
		varr := decodedTx.Message.Instructions[i]

		programPub, _ := decodedTx.Message.ResolveProgramIDIndex(varr.ProgramIDIndex)
		accounts := make([]*solana.AccountMeta, 0)
		for _, v := range varr.Accounts {
			accounts = append(accounts, metas[v])
		}
		result, err := ParseDecode(programPub, accounts, varr.Data)

		if err != nil {
			Log.Println(err)
		} else {
			vss := varr.Data.String()
			Log.Println(vss, result)

			typeKey := result.Key
			typeData := result.Data
			spew.Dump(typeKey, typeData)
			Log.Println(typeKey, typeData)
		}
	}

}

func TestParseAddLiquidity(t *testing.T) {

	t.Helper()

	maxVersion := uint64(0)
	txSig := solana.MustSignatureFromBase58("7afQYvuBqHsazkCoti2w6cMmyNzNcRQuaLmHU5ybc8J1nLTqFt8WkVLzFVYunbXsMB6YPLmgdxj636rRKcPqnBo") //add liq

	out, err := clientBlock.GetTransaction(
		context.TODO(),
		txSig,
		&rpc.GetTransactionOpts{
			Encoding:                       solana.EncodingBase64,
			MaxSupportedTransactionVersion: &maxVersion,
		},
	)
	if err != nil {
		panic(err)
	}

	txSlot := out.Slot
	txMeta := out.Meta
	txVer := out.Version
	txTran := out.Transaction

	t.Log(txSlot, txMeta, txVer, txTran)

	decodedTx, _ := solana.TransactionFromDecoder(bin.NewBinDecoder(txTran.GetBinary()))
	lookupTableAccounts := decodedTx.Message.AddressTableLookups
	for _, lookup := range lookupTableAccounts {
		lookupTable, err := addressLookupTable.GetAddressLookupTable(context.Background(), clientBlock, lookup.AccountKey)
		if err != nil {
			panic(err)
		}

		_ = decodedTx.Message.SetAddressTables(map[solana.PublicKey]solana.PublicKeySlice{
			lookup.AccountKey: lookupTable.Addresses,
		})
	}

	metas, err := decodedTx.Message.AccountMetaList()
	acc := make([]string, 0)
	for _, v := range metas {
		acc = append(acc, v.PublicKey.String())
	}

	varInners := txMeta.InnerInstructions
	t.Log(varInners)
	for i, _ := range decodedTx.Message.Instructions {
		varr := decodedTx.Message.Instructions[i]

		var insInners *[]solana.CompiledInstruction
		for j, _ := range varInners {
			if varInners[j].Index == uint16(i) {
				insInners = &varInners[j].Instructions
			}
		}
		t.Log(insInners)

		programPub, _ := decodedTx.Message.ResolveProgramIDIndex(varr.ProgramIDIndex)
		accounts := make([]*solana.AccountMeta, 0)
		for _, v := range varr.Accounts {
			accounts = append(accounts, metas[v])
		}
		result, err := ParseDecode(programPub, accounts, varr.Data, insInners)

		if err != nil {
			t.Error(err)
		} else {
			vss := varr.Data.String()
			t.Log(vss, result)

			typeKey := result.Key
			typeData := result.Data
			spew.Dump(typeKey, typeData)
			t.Log(typeKey, typeData)
		}
	}

}

func TestSwap(t *testing.T) {
	t.Helper()

	maxVersion := uint64(0)
	txSig := solana.MustSignatureFromBase58("3H5bKXUhSqQ1UPVrcGCGfoJCscT4UtFF8N5yBLSkTcHNeuRfhCRyAMMBWdFcfkdix18RT97fbovxTTgLUU6Lh9Tm") //add liq

	out, err := clientBlock.GetTransaction(
		context.TODO(),
		txSig,
		&rpc.GetTransactionOpts{
			Encoding:                       solana.EncodingBase64,
			MaxSupportedTransactionVersion: &maxVersion,
		},
	)
	if err != nil {
		panic(err)
	}

	txSlot := out.Slot
	txMeta := out.Meta
	txVer := out.Version
	txTran := out.Transaction

	t.Log(txSlot, txMeta, txVer, txTran)

	decodedTx, _ := solana.TransactionFromDecoder(bin.NewBinDecoder(txTran.GetBinary()))
	lookupTableAccounts := decodedTx.Message.AddressTableLookups

	table := make(map[solana.PublicKey]solana.PublicKeySlice)
	for _, lookup := range lookupTableAccounts {
		lookupTable, err := addressLookupTable.GetAddressLookupTable(context.Background(), dego.Client(), lookup.AccountKey)
		if err != nil {
			panic(err)
		}
		table[lookup.AccountKey] = lookupTable.Addresses
	}
	if len(table) > 0 {
		decodedTx.Message.SetAddressTables(table)
	}

	metas, err := decodedTx.Message.AccountMetaList()
	acc := make([]string, 0)
	for _, v := range metas {
		acc = append(acc, v.PublicKey.String())
	}

	varInners := txMeta.InnerInstructions
	t.Log(varInners)
	for i, _ := range decodedTx.Message.Instructions {
		varr := decodedTx.Message.Instructions[i]

		var insInners *[]solana.CompiledInstruction
		for j, _ := range varInners {
			if varInners[j].Index == uint16(i) {
				insInners = &varInners[j].Instructions
			}
		}
		t.Log(insInners)

		programPub, _ := decodedTx.Message.ResolveProgramIDIndex(varr.ProgramIDIndex)
		accounts := make([]*solana.AccountMeta, 0)
		for _, v := range varr.Accounts {
			accounts = append(accounts, metas[v])
		}
		if i == 2 {
			pr := decodedTx.Message.AccountKeys[6]
			prs := base58.Encode(pr.Bytes())
			t.Log("here break to continue", prs)
		}
		result, err := ParseDecode(programPub, metas, varr.Data, insInners, decodedTx.Message.AccountKeys)

		if err != nil {
			t.Error(err)
		} else {
			vss := varr.Data.String()
			t.Log(vss, result)

			typeKey := result.Key
			typeData := result.Data
			spew.Dump(typeKey, typeData)
			t.Log(typeKey, typeData)
		}
	}
}

func TestInnerSwap(t *testing.T) {
	t.Helper()

	maxVersion := uint64(0)
	txSig := solana.MustSignatureFromBase58("YLLExTwbXA3ZYrY3j2PRN2qqR6e7cZXT68rZYeoL8LhbympLhCnnC3f3FxqVGTEDp3jZP1bxtuSBx226zhx9QQf") //add liq

	out, err := clientBlock.GetTransaction(
		context.TODO(),
		txSig,
		&rpc.GetTransactionOpts{
			Encoding:                       solana.EncodingBase64,
			MaxSupportedTransactionVersion: &maxVersion,
		},
	)
	if err != nil {
		panic(err)
	}

	txSlot := out.Slot
	txMeta := out.Meta
	txVer := out.Version
	txTran := out.Transaction

	t.Log(txSlot, txMeta, txVer, txTran)

	decodedTx, _ := solana.TransactionFromDecoder(bin.NewBinDecoder(txTran.GetBinary()))
	lookupTableAccounts := decodedTx.Message.AddressTableLookups

	table := make(map[solana.PublicKey]solana.PublicKeySlice)
	for _, lookup := range lookupTableAccounts {
		lookupTable, err := addressLookupTable.GetAddressLookupTable(context.Background(), dego.Client(), lookup.AccountKey)
		if err != nil {
			panic(err)
		}
		table[lookup.AccountKey] = lookupTable.Addresses
	}
	if len(table) > 0 {
		decodedTx.Message.SetAddressTables(table)
	}

	metas, err := decodedTx.Message.AccountMetaList()
	acc := make([]string, 0)
	for _, v := range metas {
		acc = append(acc, v.PublicKey.String())
	}

	varInners := txMeta.InnerInstructions[0].Instructions
	t.Log(varInners)
	for i, _ := range varInners {
		varr := varInners[i]
		if i == 3 {
			t.Log("here")
		}

		var insInners *[]solana.CompiledInstruction

		// programPub, _ := decodedTx.Message.ResolveProgramIDIndex(varr.ProgramIDIndex)
		programPub := metas[varr.ProgramIDIndex].PublicKey

		programPubStr := programPub.String()
		t.Log(programPubStr)
		accounts := make([]*solana.AccountMeta, 0)
		for _, v := range varr.Accounts {
			accounts = append(accounts, metas[v])
		}
		if i == 4 {
			pr := decodedTx.Message.AccountKeys[6]
			prs := base58.Encode(pr.Bytes())
			t.Log("here break to continue", prs)
		}
		result, err := ParseDecode(programPub, metas, varr.Data, insInners, decodedTx.Message.AccountKeys)

		if err != nil {
			t.Error(err)
		} else {
			vss := varr.Data.String()
			t.Log(vss, result)

			typeKey := result.Key
			typeData := result.Data
			spew.Dump(typeKey, typeData)
			t.Log(typeKey, typeData)
		}
	}
}

func TestParseCodex(t *testing.T) {
	codex := "uuskjQnqeckwhhKCFre4W4H7kp68MUmFUsFAo4dZn4B4tp1v1komtyQfwfKo9AaPfr4DZXhUbPAHKVUFkovTQTn7tTiawziCRKrCvvDek5NQTguJdp9FaViwpvtEq23vNQ5R27Waru9W5yqj4k4KgmWjQAYACqJHtqKEWmWNcvwgHfR7PWCShhEKWirVAmXCcrWgeZxUxgZxUyBRWSKwxPddKHD7oPbsXTDCVzxFgCUpUybFmAwbD5DVdgUU9z63PBohRAdGMZjB4o4VKVHQQuiaBZeNSMk27QmXUQ3XxvwaUGEGZVLF7W8ga4d59RZYgNBEEoi9BwK4oZ5i5JrbfHVN9LqZZCgPyEDvs9Bp1grJVTGL7ywEr8SPu2FiFtkofxdPoM5mmVRnkMSZ2JqnVB6p6j95s3RQZEEF5dfZs3JuQ83TtP2RxMsTNq4wdbxKY78HoKgfHA3EXJFhUYtAvPdNKthAgcGg7NznkxE1GkAHKFHR2NhEEiBvYJnp6S1E3VntFqrKiYvhJmtrxxwdsz3oFANxQQAAr8Cx7NgWUe5uH87LGUndrUnbrqQVx8wWQnmKLTMvg7Kvjg9izx4u6Tz8FfPuHrrQsLWbDvWHSADGRpQ1EbJQXSL4snx1cqDaJDJwPS1fNEqKR7tAV5Ri4UFAejnbqbuMy65EjbBRrdhWxjWWiQ1AmPb4fij92kvuMJ5By8Q3NoBpxHd6P3Kyx31ELRnUq3gidxBHKbVgsccj5aP9JmK7U53D4nn1F11kweKMRAVFpqS5ym2kWWB86nijLkfYsnysqb986UXykWUAHsB55h13mbVaHU4ZthYYvydTUBSVca6cB1yXsrWeHZbvdta5NNATDiHYmxZgXtDfi1ysAC39v65bcTqL7RHx237qkuB2UVtSfgo8ju2NGGQsiZADQvoienxQ5gBNjAidQT5Jgj9zdpybST7BpcV9Zx6RL7a7TABuTVr1ZPcWUnmtLTAptzMsxtCDYrAsKMK7fTkWiEfQ"

	decodedBytes, err := base58.Decode(codex)

	transaction, err := solana.TransactionFromDecoder(bin.NewBinDecoder(decodedBytes))
	if err != nil {
		log.Fatalf("Failed to decode transaction: %v", err)
	}

	// 打印交易的详细信息
	fmt.Println("Transaction:", transaction)

	// 遍历并解析交易中的每个指令
	for i, instruction := range transaction.Message.Instructions {
		fmt.Printf("Instruction %d:\n", i+1)
		fmt.Println("Program ID:", instruction.ProgramIDIndex)
		fmt.Println("Accounts:", instruction.Accounts)
		fmt.Println("Data:", instruction.Data)
	}
}

func TestParseTxCodex(t *testing.T) {
	codex := "uuskjQnqeckwhhKCFre4W4H7kp68MUmFUsFAo4dZn4B4tp1v1komtyQfwfKo9AaPfr4DZXhUbPAHKVUFkovTQTn7tTiawziCRKrCvvDek5NQTguJdp9FaViwpvtEq23vNQ5R27Waru9W5yqj4k4KgmWjQAYACqJHtqKEWmWNcvwgHfR7PWCShhEKWirVAmXCcrWgeZxUxgZxUyBRWSKwxPddKHD7oPbsXTDCVzxFgCUpUybFmAwbD5DVdgUU9z63PBohRAdGMZjB4o4VKVHQQuiaBZeNSMk27QmXUQ3XxvwaUGEGZVLF7W8ga4d59RZYgNBEEoi9BwK4oZ5i5JrbfHVN9LqZZCgPyEDvs9Bp1grJVTGL7ywEr8SPu2FiFtkofxdPoM5mmVRnkMSZ2JqnVB6p6j95s3RQZEEF5dfZs3JuQ83TtP2RxMsTNq4wdbxKY78HoKgfHA3EXJFhUYtAvPdNKthAgcGg7NznkxE1GkAHKFHR2NhEEiBvYJnp6S1E3VntFqrKiYvhJmtrxxwdsz3oFANxQQAAr8Cx7NgWUe5uH87LGUndrUnbrqQVx8wWQnmKLTMvg7Kvjg9izx4u6Tz8FfPuHrrQsLWbDvWHSADGRpQ1EbJQXSL4snx1cqDaJDJwPS1fNEqKR7tAV5Ri4UFAejnbqbuMy65EjbBRrdhWxjWWiQ1AmPb4fij92kvuMJ5By8Q3NoBpxHd6P3Kyx31ELRnUq3gidxBHKbVgsccj5aP9JmK7U53D4nn1F11kweKMRAVFpqS5ym2kWWB86nijLkfYsnysqb986UXykWUAHsB55h13mbVaHU4ZthYYvydTUBSVca6cB1yXsrWeHZbvdta5NNATDiHYmxZgXtDfi1ysAC39v65bcTqL7RHx237qkuB2UVtSfgo8ju2NGGQsiZADQvoienxQ5gBNjAidQT5Jgj9zdpybST7BpcV9Zx6RL7a7TABuTVr1ZPcWUnmtLTAptzMsxtCDYrAsKMK7fTkWiEfQ"

	decodedBytes, err := base58.Decode(codex)

	// 将解码后的字节数组反序列化为 Solana 交易对象
	var transaction solana.Transaction
	transaction.UnmarshalWithDecoder(bin.NewBinDecoder(decodedBytes))
	if err != nil {
		log.Fatalf("Failed to decode transaction: %v", err)
	}

	walletPrivateKey := ""
	signer, err := types.AccountFromBase58(walletPrivateKey)

	// 获取 rent-exempt 最低余额（计算创建 token account 需要的最小 SOL）

	client := rpc.New("https://api.mainnet-beta.solana.com")

	rentExemption, err := client.GetMinimumBalanceForRentExemption(context.Background(), token.TokenAccountSize, rpc.CommitmentFinalized)
	if err != nil {
		log.Fatalf("Failed to get rent exemption: %v", err)
	}

	tokenProgramID := solana.MustPublicKeyFromBase58("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA")

	// 添加创建账户的指令 (System Program: Create Account)
	createAccountInstruction := system.NewCreateAccountInstruction(
		rentExemption,                      // rent exempt 最低余额
		token.TokenAccountSize,             // SPL token account 所需空间大小
		tokenProgramID,                     // SPL token program ID
		transaction.Message.AccountKeys[0], // 资助者账户（通常是交易发起者）
		solana.PublicKey(signer.PublicKey), // 要创建的 token account 公钥
	)

	// 添加初始化 token 账户的指令 (Token Program: Initialize Account)
	mintPubKey := solana.MustPublicKeyFromBase58("TokenMintPublicKey") // 代币 mint 地址
	ownerPubKey := transaction.Message.AccountKeys[0]                  // token account 的所有者

	initializeTokenAccountInstruction := token.InitializeAccount(
		token.InitializeAccountParam{
			Account: signer.PublicKey,              // 要初始化的 token account
			Mint:    common.PublicKey(mintPubKey),  // SPL Token mint
			Owner:   common.PublicKey(ownerPubKey), // Token account owner
		},
	)

	createData, _ := createAccountInstruction.Build().Data()

	compiledCreateAccountInstruction := solana.CompiledInstruction{
		ProgramIDIndex: 0,              // 根据 AccountKeys 中 programID 的索引，这里需要是 System Program 的索引
		Accounts:       []uint16{0, 1}, // 参与的账户索引：资助者和新账户
		Data:           createData,     // 指令的数据
	}

	// 手动构建 initializeTokenAccountInstruction 的 CompiledInstruction
	compiledInitializeTokenAccountInstruction := solana.CompiledInstruction{
		ProgramIDIndex: 1,                                      // 根据 AccountKeys 中 tokenProgramID 的索引
		Accounts:       []uint16{0, 1, 2},                      // 参与的账户索引：新账户、mint、owner
		Data:           initializeTokenAccountInstruction.Data, // 指令的数据
	}

	// 将新的指令添加到交易中
	transaction.Message.Instructions = append(
		transaction.Message.Instructions,
		compiledCreateAccountInstruction, // 添加创建账户指令
		compiledInitializeTokenAccountInstruction,
	)

	// 重新计算消息哈希并签署交易

	messageHash, _ := transaction.Message.MarshalBinary()
	signature := solana.SignatureFromBytes(signer.Sign(messageHash))
	if err != nil {
		log.Fatalf("Failed to sign message hash: %v", err)
	}

	// 添加签名到交易中
	transaction.Signatures = append(transaction.Signatures, signature)

	// 序列化并发送新的交易
	newTxBytes, err := transaction.MarshalBinary()
	if err != nil {
		log.Fatalf("Failed to serialize new transaction: %v", err)
	}

	// 打印新签名的交易
	fmt.Println("New Transaction Base58:", base58.Encode(newTxBytes))
}
