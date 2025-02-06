package rpc

import (
	"encoding/json"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/powershitxyz/SolanaProbe/dego"
	"github.com/powershitxyz/SolanaProbe/sys"
)

var slotQueue = sys.NewSlotQueue()

var logger = sys.Logger
var lastCheck int64
var subscribeDone = make(chan struct{})

func ReadSlotQueue() *sys.SlotQueue {
	return slotQueue
}

func InitEssential() {
	if wsRpc := conf.Chain.WsRpc; wsRpc == "" {
		logger.Fatalln("error config websocket rpc")
	}

	go connectAndSubscribe()

	for range subscribeDone {
		logger.Println("Reconnecting...")
		time.Sleep(3 * time.Second)
		go connectAndSubscribe()
	}
}

func connectAndSubscribe() {
	u, err := url.Parse(conf.Chain.WsRpc)
	atomic.StoreInt64(&lastCheck, time.Now().Unix()) // initialized connection time
	if err != nil {
		logger.Fatal("Error parsing URL:", err)
	}
	logger.Printf("Connecting to %s", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		logger.Printf("Failed to connect to WebSocket: %v", err)
		triggerReconnect()
		return
	}
	defer conn.Close()

	conn.SetPongHandler(func(appData string) error {
		atomic.StoreInt64(&lastCheck, time.Now().Unix())
		return nil
	})

	pingTicker := time.NewTicker(5 * time.Second)
	defer pingTicker.Stop()

	go sendPing(conn, pingTicker)

	notificationChannel := make(chan dego.Notification, 100)
	go processNotifications(notificationChannel)

	timeoutTicker := time.NewTicker(30 * time.Second)
	defer timeoutTicker.Stop()

	go readMessages(conn, notificationChannel)

	WatchSlotSubscribe(conn)

	for range timeoutTicker.C {
		if time.Now().Unix()-atomic.LoadInt64(&lastCheck) > 30 {
			logger.Info("Timeout triggered, reconnecting...")
			triggerReconnect()
			return
		}
	}
}

func sendPing(conn *websocket.Conn, ticker *time.Ticker) {
	for range ticker.C {
		if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			logger.Println("Ping failed:", err)
			triggerReconnect()
			return
		}
	}
}

func readMessages(conn *websocket.Conn, notificationChannel chan dego.Notification) {
	for {
		_, message, err := conn.ReadMessage()
		if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
			logger.Println("WebSocket closed normally:", err)
			return
		} else if err != nil {
			logger.Println("WebSocket read error:", err)
			triggerReconnect()
			return
		}

		var response map[string]interface{}
		if err := json.Unmarshal(message, &response); err != nil {
			atomic.StoreInt64(&lastCheck, time.Now().Unix())
			logger.Println("Unmarshal error:", err)
			continue
		}

		method, ok := response["method"].(string)
		if !ok {
			atomic.StoreInt64(&lastCheck, time.Now().Unix())
			logger.Printf("Method not found in response: [%s], ReadMsg: %v", method, response)
			continue
		}
		atomic.StoreInt64(&lastCheck, time.Now().Unix())

		obj, err := dego.RouteNotification(method, response)
		if err == nil {
			notificationChannel <- obj
		}
	}
}

func processNotifications(notificationChannel chan dego.Notification) {
	batch := make([]dego.Notification, 0, 3)
	tick := time.NewTicker(1 * time.Second)
	defer tick.Stop()

	for {
		select {
		case notification := <-notificationChannel:
			batch = append(batch, notification)
			logger.Println("New data pushed", notification)
			if len(batch) >= 10 {
				processBatch(batch)
				batch = batch[:0]
			}
		case <-tick.C:
			if len(batch) > 0 {
				processBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

func triggerReconnect() {
	select {
	case subscribeDone <- struct{}{}:
	default:
	}
}
