package rpc

import (
	"encoding/json"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/powershitxyz/SolanaProbe/config"
	"github.com/powershitxyz/SolanaProbe/dego"
	"github.com/powershitxyz/SolanaProbe/sys"
)

var slotQueue = sys.NewSlotQueue()
var unqiueSlotMap = make(map[uint64]bool)
var slotMapLock = sync.Mutex{}
var slotEOF = make([]uint64, 0)
var slotEOFLock = sync.Mutex{}
var slotEOFFlag = false

var logger = sys.Logger
var conf = config.GetConfig()

func ReadSloqQueue() *sys.SlotQueue {
	return slotQueue
}

var SubscribeDone = make(chan bool)

func InitEssential() {
	if wsRpc := conf.Chain.WsRpc; len(wsRpc) == 0 {
		logger.Fatalln("error config websocket rpc")
	}

	//触发重连逻辑 测试用
	//go time.AfterFunc(1*time.Minute, func() {
	//	SubscribeDone <- true
	//})
	go connectAndSubscribe(SubscribeDone)

	for range SubscribeDone {
		logger.Println("Reconnecting...")
		time.Sleep(3 * time.Second)
		go connectAndSubscribe(SubscribeDone)

	}
}

func connectAndSubscribe(done chan bool) {
	u, err := url.Parse(conf.Chain.WsRpc)
	var lastCheck = time.Now().Unix()
	if err != nil {
		logger.Fatal("Error parsing URL:", err)
	}
	logger.Printf("connecting to %s", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		logger.Printf("Failed to connect to WebSocket: %v", err)
		done <- true
		return
	}
	defer func() {
		err2 := conn.Close()
		if err2 != nil {
			logger.Printf("Failed to close WebSocket connection: %v", err2)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	notificationChannel := make(chan dego.Notification, 100)

	go func() {
		batch := make([]dego.Notification, 0, 3)
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case notification := <-notificationChannel:
				batch = append(batch, notification)
				if len(batch) >= 10 {
					// processBatch(batch)
					batch = batch[:0]
				}
			case <-ticker.C:
				if len(batch) > 0 {
					// processBatch(batch)
					batch = batch[:0]
				}
			}
		}
	}()

	var once sync.Once

	ticker1 := time.NewTicker(6 * time.Second)
	defer ticker1.Stop()
	// Listen for messages
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				logger.Println("read:", err)
				once.Do(func() { done <- true })
				return
			}

			var response map[string]interface{}
			err = json.Unmarshal(message, &response)
			if err != nil {
				atomic.StoreInt64(&lastCheck, time.Now().Unix())
				logger.Println("unmarshal:", err)
				continue
			}

			method, ok := response["method"].(string)
			if !ok {
				atomic.StoreInt64(&lastCheck, time.Now().Unix())
				slotEOFFlag = true
				logger.Printf("method not found in response:[%s],ReadMsg:%v", method, response)
				continue
			}
			atomic.StoreInt64(&lastCheck, time.Now().Unix())
			obj, err := dego.RouteNotification(method, response)
			if err == nil {
				notificationChannel <- obj
			}
		}
	}()

	for {
		select {
		case <-done:
			logger.Println("Connection closed, reconnecting...")
			once.Do(func() { done <- true })
			return // 退出当前函数以触发重连逻辑
		case t := <-ticker.C:
			err := conn.WriteMessage(websocket.PingMessage, []byte(t.String()))
			if err != nil {
				logger.Println("write ping:", err)
				once.Do(func() { done <- true })
				return // 退出当前函数以触发重连逻辑
			}
		case t1 := <-ticker1.C:
			if time.Now().Unix()-atomic.LoadInt64(&lastCheck) > 11 {
				once.Do(func() { done <- true })
				logger.Info("定时器触发重连", t1.String())
				return
			}
		}
	}
}
