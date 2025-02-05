package dego

type LogsSubscribeRequest struct {
	SolanaRequest `json:",inline"`
}

type LogsNotification struct {
	Jsonrpc string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  LogsNotificationParams `json:"params"`
}

type LogsNotificationParams struct {
	Result       LogsNotificationResult `json:"result"`
	Subscription int                    `json:"subscription"`
}

type LogsNotificationResult struct {
	Context LogsNotificationContext `json:"context"`
	Value   LogsNotificationValue   `json:"value"`
}

type LogsNotificationContext struct {
	Slot uint64 `json:"slot"`
}

type LogsNotificationValue struct {
	Signature string   `json:"signature"`
	Err       string   `json:"err"`
	Logs      []string `json:"logs"`
}
