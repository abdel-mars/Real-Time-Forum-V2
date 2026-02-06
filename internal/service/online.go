package service

import (
	"sync"
)

var (
	onlineUsers   = make(map[string]bool) // username -> true
	onlineUsersMu sync.Mutex
)

func AddOnlineUser(username string) {
	onlineUsersMu.Lock()
	defer onlineUsersMu.Unlock()
	onlineUsers[username] = true
	//BroadcastUsers()
}

func RemoveOnlineUser(username string) {
	onlineUsersMu.Lock()
	defer onlineUsersMu.Unlock()
	delete(onlineUsers, username)
	//BroadcastUsers()
}

func GetOnlineUsers() []string {
	onlineUsersMu.Lock()
	defer onlineUsersMu.Unlock()
	users := make([]string, 0, len(onlineUsers))
	for u := range onlineUsers {
		users = append(users, u)
	}
	return users
}

