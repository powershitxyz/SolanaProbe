package dego

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/powershitxyz/SolanaProbe/sys"
)

type Notification interface {
	GetMethod() string
}

type Subscriber interface {
	BuildSubscribe(iconn *websocket.Conn, params ...interface{}) SolanaRequest
}

type SolanaRequest struct {
	Subscriber Subscriber    `json:"-"`
	JsonRPC    string        `json:"jsonrpc"`
	ID         int           `json:"id"`
	Method     string        `json:"method"`
	Params     []interface{} `json:"params"`
}

type ReqParamGeneral struct {
	Commitment string `json:"commitment"`
	Encoding   string `json:"encoding"`
}

type SolanaResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
}

func (r SolanaResponse) GetMethod() string {
	return r.Method
}

func (SolanaRequest) startSubscribe(conn *websocket.Conn, v interface{}) {
	b, _ := json.Marshal(v)
	fmt.Println("startSubscribe : " + string(b))
	err := conn.WriteMessage(websocket.TextMessage, b)
	if err != nil {
		sys.Logger.Fatalf("Failed to send program subscription request: %v", err)
	}
}

func (t *SolanaRequest) BuildSubscribe(conn *websocket.Conn) {

	t.startSubscribe(conn, t)
}
