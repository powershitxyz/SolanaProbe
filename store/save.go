package store

import (
	"encoding/json"

	"github.com/powershitxyz/SolanaProbe/database"
	"github.com/powershitxyz/SolanaProbe/model"
	"github.com/powershitxyz/SolanaProbe/pub"
	"github.com/shopspring/decimal"
)

func BatchSaveTokenFlow(tokenFlows []model.TokenFlow, size int, RetryBlock bool) {
	//_ = db.CreateInBatches(tokenFlows, size)
	db := database.GetDb()
	for _, flow := range tokenFlows {
		if flow.Price.Cmp(decimal.Zero) <= 0 {
			continue
		}
		if !RetryBlock {
			jsonData, err := json.Marshal(flow)
			if err != nil {
				continue
			}
			database.PublishS(pub.TokenFlowTopic, jsonData)
		} else {
			reFLow := &flow
			reFLow.Retry = 1
			jsonData, err := json.Marshal(reFLow)
			if err != nil {
				continue
			}
			database.PublishS(pub.TokenFlowTopic, jsonData)
			db.CreateInBatches(tokenFlows, size)
		}

	}
}
