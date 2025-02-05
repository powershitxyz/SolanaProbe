package dego

import (
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type RawTransaction struct {
	Slot        uint64
	Transaction *solana.Transaction
	Meta        *rpc.TransactionMeta
}
