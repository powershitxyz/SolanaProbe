package dego

import (
	"encoding/json"
	"fmt"

	"github.com/powershitxyz/SolanaProbe/database"
	"github.com/powershitxyz/SolanaProbe/sys"
)

var log = sys.Logger
var db = database.GetDb()

type NotificationHandler func(notification interface{}) (Notification, error)

var handlers = map[string]NotificationHandler{
	"programNotification":      handleProgramNotification,
	"slotsUpdatesNotification": handleSlotUpdatesNotification,
	"slotNotification":         handleSlotSubscribeNotification,
	"blockNotification":        handleBlockSubscribeNotification,
	// Add other handlers for different types of notifications
}

func RouteNotification(notificationType string, notification interface{}) (Notification, error) {
	if handler, exists := handlers[notificationType]; exists {
		obj, err := handler(notification)
		if err != nil {
			log.Printf("Failed to handle %s notification: %v\n", notificationType, err)
		}
		return obj, err
	}
	return nil, fmt.Errorf("unknown notification type %s", notificationType)
}

func handleProgramNotification(notification interface{}) (Notification, error) {
	var programNotification ProgramNotification
	data, err := json.Marshal(notification)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal notification: %v", err)
	}
	err = json.Unmarshal(data, &programNotification)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal program notification: %v", err)
	}
	return programNotification, nil
}

func handleSlotUpdatesNotification(notification interface{}) (Notification, error) {
	var slotNotification SlotUpdateNotification
	data, err := json.Marshal(notification)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal notification: %v", err)
	}
	err = json.Unmarshal(data, &slotNotification)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal program notification: %v", err)
	}
	return slotNotification, nil
}

func handleSlotSubscribeNotification(notification interface{}) (Notification, error) {
	var slotNotification SlotNotification
	data, err := json.Marshal(notification)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal notification: %v", err)
	}
	err = json.Unmarshal(data, &slotNotification)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal program notification: %v", err)
	}
	return slotNotification, nil
}
func handleBlockSubscribeNotification(notification interface{}) (Notification, error) {
	var slotNotification BlockSubNotification
	data, err := json.Marshal(notification)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal notification: %v", err)
	}
	err = json.Unmarshal(data, &slotNotification)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal program notification: %v", err)
	}
	return slotNotification, nil
}
