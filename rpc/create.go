package rpc

import (
	"sync"

	"github.com/powershitxyz/SolanaProbe/dego"
)

var id int
var idMapping map[int]interface{}
var mutex sync.Mutex

func init() {
	id = 0
	idMapping = make(map[int]interface{})
}

func increment() int {
	var idTemp int
	mutex.Lock()
	{
		id++
		idTemp = id
	}
	mutex.Unlock()
	return idTemp

}

func NewSolanaRequest(sub dego.Subscriber, jsonRPC string, id int, method string, params []interface{}) *dego.SolanaRequest {
	return &dego.SolanaRequest{
		Subscriber: sub,
		JsonRPC:    jsonRPC,
		ID:         id,
		Method:     method,
		Params:     params,
	}
}

// Factory method for ProgramSubscribeRequest
func NewProgramSubscribeRequest(params []interface{}) *dego.ProgramSubscribeRequest {
	idTemp := increment()
	idMapping[idTemp] = "programSubscribe"
	return &dego.ProgramSubscribeRequest{
		SolanaRequest: dego.SolanaRequest{
			JsonRPC: "2.0",
			ID:      idTemp,
			Method:  "programSubscribe",
			Params:  params,
		},
	}
}

func NewSlotSubscribeRequest(params []interface{}) *dego.SlotSubscribeRequest {

	idTemp := increment()
	idMapping[idTemp] = "slotSubscribe"
	return &dego.SlotSubscribeRequest{
		SolanaRequest: dego.SolanaRequest{
			JsonRPC: "2.0",
			ID:      idTemp,
			Method:  "slotSubscribe",
			Params:  params,
		},
	}
}

func NewSlotUpdatesSubscribeRequest(params []interface{}) *dego.SlotSubscribeRequest {
	idTemp := increment()
	idMapping[idTemp] = "slotsUpdatesSubscribe"
	return &dego.SlotSubscribeRequest{
		SolanaRequest: dego.SolanaRequest{
			JsonRPC: "2.0",
			ID:      idTemp,
			Method:  "slotsUpdatesSubscribe",
			Params:  params,
		},
	}
}

func NewBlockSubscribeRequest(params []interface{}) *dego.SlotSubscribeRequest {
	idTemp := increment()
	idMapping[idTemp] = "blockSubscribe"
	return &dego.SlotSubscribeRequest{
		SolanaRequest: dego.SolanaRequest{
			JsonRPC: "2.0",
			ID:      idTemp,
			Method:  "blockSubscribe",
			Params:  params,
		},
	}
}

// Factory method for LogsSubscribeRequest
func NewLogsSubscribeRequest(id int, params []interface{}) *dego.LogsSubscribeRequest {
	return &dego.LogsSubscribeRequest{
		SolanaRequest: dego.SolanaRequest{
			JsonRPC: "2.0",
			ID:      id,
			Method:  "logsSubscribe",
			Params:  params,
		},
	}
}
