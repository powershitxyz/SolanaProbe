package rpc

import (
	"github.com/gagliardetto/solana-go"
	"github.com/powershitxyz/SolanaProbe/config"
)

var conf = config.GetConfig()

type TokenAccountInfo struct {
	Mint            solana.PublicKey  `json:"mint"`
	Owner           solana.PublicKey  `json:"owner"`
	Amount          uint64            `json:"amount"`
	Delegate        *solana.PublicKey `json:"delegate,omitempty"`
	State           uint8             `json:"state"`
	IsNative        bool              `json:"is_native"`
	DelegatedAmount uint64            `json:"delegated_amount"`
	CloseAuthority  *solana.PublicKey `json:"close_authority,omitempty"`
}
