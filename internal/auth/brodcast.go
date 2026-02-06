package auth

import (
	"encoding/json"
)

// Define an interface for broadcasting without importing handler
type Broadcaster interface {
	BroadcastToAllRooms([]byte)
}

// Global hub instance - use interface to avoid import cycle
var GlobalHub Broadcaster

func BroadcastUsers() {
	if GlobalHub == nil {
		return // Hub not initialized yet
	}

	users := GetOnlineUsersList()

	payload, _ := json.Marshal(map[string]interface{}{
		"type":  "users_update",
		"users": users,
	})

	GlobalHub.BroadcastToAllRooms(payload)
}

// BroadcastUserJoined sends a user joined event
func BroadcastUserJoined(username string) {
	if GlobalHub == nil {
		return
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"type":     "user_joined",
		"username": username,
	})

	GlobalHub.BroadcastToAllRooms(payload)
}

// BroadcastUserLeft sends a user left event
func BroadcastUserLeft(username string) {
	if GlobalHub == nil {
		return
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"type":     "user_left",
		"username": username,
	})

	GlobalHub.BroadcastToAllRooms(payload)
}
