package dego

import "encoding/json"

type ProgramSubscribeRequest struct {
	SolanaRequest `json:",inline"`
}

type ProgramNotification struct {
	SolanaResponse
	Params ProgramParams `json:"params"`
}

type ProgramParams struct {
	Result       ProgramResult `json:"result"`
	Subscription int           `json:"subscription"`
}

type ProgramResult struct {
	Context ProgramContext `json:"context"`
	Value   ProgramValue   `json:"value"`
}

type ProgramContext struct {
	Slot uint64 `json:"slot"`
}

type ProgramValue struct {
	Pubkey  string         `json:"pubkey"`
	Account ProgramAccount `json:"account"`
}

type ProgramAccount struct {
	Lamports   uint64      `json:"lamports"`
	Data       ProgramData `json:"data"`
	Owner      string      `json:"owner"`
	Executable bool        `json:"executable"`
	RentEpoch  json.Number `json:"rentEpoch"`
}

type ProgramData struct {
	Program string        `json:"program"`
	Parsed  ProgramParsed `json:"parsed"`
	Space   int           `json:"space"`
}

type ProgramParsed struct {
	Info ProgramInfo `json:"info"`
	Type string      `json:"type"`
}

type ProgramInfo struct {
	IsNative          bool        `json:"isNative,omitempty"`
	Mint              string      `json:"mint,omitempty"`
	Owner             string      `json:"owner,omitempty"`
	State             string      `json:"state,omitempty"`
	TokenAmount       TokenAmount `json:"tokenAmount,omitempty"`
	Decimals          int         `json:"decimals,omitempty"`
	FreezeAuthority   string      `json:"freezeAuthority,omitempty"`
	IsInitialized     bool        `json:"isInitialized,omitempty"`
	MintAuthority     string      `json:"mintAuthority,omitempty"`
	Supply            string      `json:"supply,omitempty"`
	Source            string      `json:"source,omitempty"`
	Destination       string      `json:"destination,omitempty"`
	AmountTransferred string      `json:"amountTransferred,omitempty"`
}

type TokenAmount struct {
	Amount         string  `json:"amount"`
	Decimals       int     `json:"decimals"`
	UiAmount       float64 `json:"uiAmount"`
	UiAmountString string  `json:"uiAmountString"`
}

func (p ProgramNotification) GetMethod() string {
	return p.Method
}
