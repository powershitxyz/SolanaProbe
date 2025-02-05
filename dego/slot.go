package dego

import "github.com/gagliardetto/solana-go/rpc"

type BlockNotification struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  struct {
		Result struct {
			Context struct {
				Slot uint64 `json:"slot"`
			} `json:"context"`
			Value struct {
				Slot         uint64      `json:"slot"`
				Blockhash    string      `json:"blockhash,omitempty"`
				Parent       uint64      `json:"parentSlot,omitempty"`
				Transactions []string    `json:"transactions,omitempty"`
				Err          interface{} `json:"err,omitempty"`
			} `json:"value"`
		} `json:"result"`
		Subscription int `json:"subscription"`
	} `json:"params"`
}

type SlotSubscribeRequest struct {
	SolanaRequest `json:",inline"`
}

/*** Slot Notification ***/
type SlotNotification struct {
	SolanaResponse
	Params SlotParams `json:"params"`
}

type SlotParams struct {
	Result       SlotResult `json:"result"`
	Subscription uint64     `json:"subscription"`
}

type SlotResult struct {
	Parent uint64 `json:"parent"`
	Root   uint64 `json:"root"`
	Slot   uint64 `json:"slot"`
	Status string `json:"status"`
	Leader string `json:"leader"`
}

/*** Slot Update Notification ***/
type SlotUpdateNotification struct {
	SolanaResponse
	Params SlotUpdateParams `json:"params"`
}

type SlotUpdateParams struct {
	Result       SlotUpdateResult `json:"result"`
	Subscription uint64           `json:"subscription"`
}

type SlotUpdateResult struct {
	Parent    uint64 `json:"parent"`
	Slot      uint64 `json:"slot"`
	Root      uint64 `json:"root"`
	Type      string `json:"type"`
	Timestamp uint64 `json:"timestamp"`
}

type BlockSubNotification struct {
	SolanaResponse
	Params BlockSubParams `json:"params"`
}
type BlockSubParams struct {
	Result       BlockSubContentResult `json:"result"`
	Subscription uint64                `json:"subscription"`
}
type BlockSubContentResult struct {
	Context ProgramContext      `json:"context"`
	Value   BlockSubValueResult `json:"value"`
}
type BlockSubValueResult struct {
	Slot  uint64             `json:"slot"`
	Block rpc.GetBlockResult `json:"block"`
	Err   struct{}           `json:"err"`
}
