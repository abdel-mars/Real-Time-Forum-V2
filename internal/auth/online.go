package auth

import (
	db "forum/internal/db"
	"encoding/json"
	"sync"

)

var (
	onlineUsers        = make(map[string]bool) // username -> true
	onlineUsersMu      sync.Mutex
	connectionCounts   = make(map[string]int)
	connectionCountsMu sync.Mutex
)

// ---------------------------
// Connection tracking helpers
// ---------------------------
func AddUserConnection(username string) {
	if username == "" {
		return
	}
	connectionCountsMu.Lock()
	connectionCounts[username]++
	shouldSignal := connectionCounts[username] == 1
	connectionCountsMu.Unlock()
	if shouldSignal {
		AddOnlineUser(username)
	}
}

func RemoveUserConnection(username string) {
	if username == "" {
		return
	}
	connectionCountsMu.Lock()
	cnt := connectionCounts[username]
	if cnt <= 0 {
		connectionCountsMu.Unlock()
		return
	}
	connectionCounts[username] = cnt - 1
	shouldSignal := connectionCounts[username] == 0
	connectionCountsMu.Unlock()
	if shouldSignal {
		RemoveOnlineUser(username)
	}
}

func AddOnlineUser(username string) {
	if username == "" {
		return
	}
	onlineUsersMu.Lock()
	if onlineUsers[username] {
		onlineUsersMu.Unlock()
		return
	}
	onlineUsers[username] = true
	onlineUsersMu.Unlock()
	// Update database
	if userID, err := db.GetUserIDByUsername(username); err == nil {
		db.SetUserOnline(userID)
	}
	// Broadcast user joined event
	if GlobalHub != nil {
		message := map[string]interface{}{
			"type":     "user_joined",
			"username": username,
		}
		if msgBytes, err := json.Marshal(message); err == nil {
			GlobalHub.BroadcastToAllRooms(msgBytes)
		}
	}
}

func RemoveOnlineUser(username string) {
	if username == "" {
		return
	}
	onlineUsersMu.Lock()
	wasOnline := onlineUsers[username]
	if wasOnline {
		delete(onlineUsers, username)
	}
	onlineUsersMu.Unlock()
	if !wasOnline {
		return
	}
	connectionCountsMu.Lock()
	delete(connectionCounts, username)
	connectionCountsMu.Unlock()
	// Update database
	if userID, err := db.GetUserIDByUsername(username); err == nil {
		db.SetUserOffline(userID)
	}
	// Broadcast user left event
	if GlobalHub != nil {
		message := map[string]interface{}{
			"type":     "user_left",
			"username": username,
		}
		if msgBytes, err := json.Marshal(message); err == nil {
			GlobalHub.BroadcastToAllRooms(msgBytes)
		}
	}
}

func GetOnlineUsersList() []string {
	onlineUsersMu.Lock()
	defer onlineUsersMu.Unlock()
	users := make([]string, 0, len(onlineUsers))
	for u := range onlineUsers {
		users = append(users, u)
	}
	return users
}

// LoadOnlineUsersFromDB loads online users from database into memory
func LoadOnlineUsersFromDB() error {
	usernames, err := db.GetOnlineUsers()
	if err != nil {
		return err
	}
	onlineUsersMu.Lock()
	defer onlineUsersMu.Unlock()
	for _, username := range usernames {
		onlineUsers[username] = true
	}
	return nil
}
