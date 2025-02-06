package rpc

import (
	"github.com/gagliardetto/solana-go"
	"github.com/gorilla/websocket"
)

func WatchSlotUpdatesSubscribe(conn *websocket.Conn) {
	NewSlotUpdatesSubscribeRequest([]interface{}{
		// PprogramId.TokenProgramId,
		// map[string]interface{}{
		// 	"commitment": "finalized",
		// 	"encoding":   "jsonParsed",
		// },
	}).BuildSubscribe(conn)
}

func WatchSlotSubscribe(conn *websocket.Conn) {
	NewSlotSubscribeRequest([]interface{}{
		// PprogramId.TokenProgramId,
		// map[string]interface{}{
		// 	"commitment": "finalized",
		// 	"encoding":   "jsonParsed",
		// },
	}).BuildSubscribe(conn)
}
func WatchBlockSubscribe(conn *websocket.Conn) {
	NewBlockSubscribeRequest([]interface{}{
		"all",
		map[string]interface{}{
			"commitment":                     "confirmed",
			"encoding":                       solana.EncodingBase64,
			"showRewards":                    false,
			"transactionDetails":             "full",
			"maxSupportedTransactionVersion": 0,
		},
	}).BuildSubscribe(conn)
}
