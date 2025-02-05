package dego

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type RPCRequest struct {
	Jsonrpc string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type RPCResponse struct {
	Jsonrpc string           `json:"jsonrpc"`
	ID      int              `json:"id"`
	Result  *json.RawMessage `json:"result"`
	Error   *RPCError        `json:"error"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Transaction struct {
	Meta        interface{} `json:"meta"`
	Transaction interface{} `json:"transaction"`
}

type Block struct {
	Blockhash         string        `json:"blockhash"`
	PreviousBlockhash string        `json:"previousBlockhash"`
	ParentSlot        int           `json:"parentSlot"`
	Transactions      []Transaction `json:"transactions"`
}

func GetBlock(slot uint64) (*Block, error) {
	solanaRPCURL := conf.Chain.GetRpc()

	request := RPCRequest{
		Jsonrpc: "2.0",
		ID:      1,
		Method:  "getBlock",
		Params: []interface{}{(slot), map[string]interface{}{
			"encoding":                       "json",
			"maxSupportedTransactionVersion": 0,
		}},
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(solanaRPCURL[0], "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rpcResponse RPCResponse
	if err := json.Unmarshal(body, &rpcResponse); err != nil {
		return nil, err
	}

	if rpcResponse.Error != nil {
		return nil, fmt.Errorf("%d:%s", rpcResponse.Error.Code, rpcResponse.Error.Message)
	}

	var block Block
	if err := json.Unmarshal(*rpcResponse.Result, &block); err != nil {
		return nil, err
	}

	return &block, nil
}
