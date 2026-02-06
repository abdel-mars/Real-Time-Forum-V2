package handler

import (
	auth "forum/internal/auth"
	db "forum/internal/db"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strings"
	"github.com/gorilla/websocket"
)

type Hub struct {
	Rooms             map[string]*room
	globalUsers       map[string]*user
	notificationUsers map[string]*user
}

// Create a new Hub
func NewHub() *Hub {
	h := &Hub{
		Rooms:             make(map[string]*room),
		globalUsers:       make(map[string]*user),
		notificationUsers: make(map[string]*user),
	}
	go h.run()
	return h
}

func (h *Hub) run() {
}

func (h *Hub) broadcastUsersToAllRooms() {
}

func (h *Hub) broadcastGlobalUsers() {
}

func (h *Hub) BroadcastToAllRooms(message []byte) {
	for _, room := range h.Rooms {
		select {
		case room.forward <- message:
		default:
		}
	}
}

func CreatePrivateRoomName(user1, user2 string) string {
	users := []string{user1, user2}
	sort.Strings(users)
	return "private_" + strings.Join(users, "_")
}

func (h *Hub) GetRoom(name string) *room {
	if r, ok := h.Rooms[name]; ok {
		return r
	}
	r := NewRoom(name, h)
	h.Rooms[name] = r
	go r.Run()
	return r
}

// Room stores users and channels
type room struct {
	users   map[*user]bool
	join    chan *user
	leave   chan *user
	forward chan []byte
	name    string
	hub     *Hub
}

// Create a new Room with name
func NewRoom(name string, hub *Hub) *room {
	return &room{
		users:   make(map[*user]bool),
		join:    make(chan *user),
		leave:   make(chan *user),
		forward: make(chan []byte),
		name:    name,
		hub:     hub,
	}
}

// ChatMessageData represents the structure of chat messages
type ChatMessageData struct {
	Type      string `json:"type"`
	Name      string `json:"name"`
	Message   string `json:"message"`
	SenderID  int    `json:"sender_id,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

// Run handles all room events
func (r *room) Run() {
	for {
		select {
		case u := <-r.join:
			r.users[u] = true
			log.Printf("User %s joined room %s (now %d users in room)", u.name, r.name, len(r.users))
			for otherUser := range r.users {
				log.Printf("  - User in room: %s", otherUser.name)
			}

		case u := <-r.leave:
			if _, ok := r.users[u]; ok {
				delete(r.users, u)
				close(u.recieve)
				log.Printf("User %s left room %s", u.name, r.name)
			}

		case msg := <-r.forward:
			var chatMsg ChatMessageData
			if err := json.Unmarshal(msg, &chatMsg); err == nil {
				log.Printf(" Processing message in room %s: %+v", r.name, chatMsg)

				switch chatMsg.Type {
				case "message":
					if strings.HasPrefix(r.name, "private_") {
						usernames := strings.Split(strings.TrimPrefix(r.name, "private_"), "_")
						if len(usernames) == 2 {
							// Get sender ID - prioritize sender_id from message
							var senderID int
							if chatMsg.SenderID > 0 {
								senderID = chatMsg.SenderID
							} else {
								senderID = r.getUserIDByUsername(chatMsg.Name)
							}

							// Determine receiver
							receiverName := usernames[0]
							if receiverName == chatMsg.Name {
								receiverName = usernames[1]
							}
							receiverID := r.getUserIDByUsername(receiverName)

							if senderID <= 0 || receiverID <= 0 {
								// Still forward the message even if save fails
								break
							}

							// Save message to database
							err := db.SaveChatMessage(senderID, receiverID, chatMsg.Message)
							if err == nil {
								// IMPORTANT: Retrieve the saved message to get the created_at timestamp
								messages, err := db.GetChatMessages(senderID, receiverID, 1, 0)
								if err == nil && len(messages) > 0 {
									// Update the message with the actual created_at from database
									chatMsg.CreatedAt = messages[0].CreatedAt
									// Re-marshal with updated timestamp
									if updatedMsg, err := json.Marshal(chatMsg); err == nil {
										msg = updatedMsg
									}
								}

								// Send notification to receiver if they're connected for notifications
								if notifUser, exists := r.hub.notificationUsers[receiverName]; exists {
									select {
									case notifUser.recieve <- msg:
									default:
									}
								}
							}
						}
					}
				case "typing", "stop_typing":
					log.Printf("Typing indicator: %s from %s", chatMsg.Type, chatMsg.Name)
				}
			}

			// Forward message to all users in the room
			log.Printf(" DEBUG: Room %s has %d users", r.name, len(r.users))
			var senderName string
			if chatMsg.Type == "message" {
				senderName = chatMsg.Name
			}
			for usr := range r.users {
				// Skip sending the message back to the sender to prevent duplicates
				if usr.name == senderName {
					log.Printf(" DEBUG: Skipping sender %s", usr.name)
					continue
				}
				log.Printf(" DEBUG: Forwarding to user %s, message: %s", usr.name, string(msg))
				select {
				case usr.recieve <- msg:
					log.Printf(" Message forwarded to user %s", usr.name)
				default:
					log.Printf(" User %s channel full, message dropped", usr.name)
				}
			}
		}
	}
}

// Helper method to get user ID by username from database
func (r *room) getUserIDByUsername(username string) int {
	userID, err := db.GetUserIDByUsername(username)
	if err != nil {
		log.Printf("Error getting user ID for %s: %v", username, err)
		return 0
	}
	return userID
}

// broadcastUsers sends the current online users in the room
func (r *room) broadcastUsers() {
	// No longer broadcasting - users stay online based on session
}

// Constants for WebSocket buffer sizes
const (
	socketBuffersize  = 1024
	messageBuffersize = 256
)

// WebSocket upgrader
var upgrader = &websocket.Upgrader{
	ReadBufferSize:  socketBuffersize,
	WriteBufferSize: messageBuffersize,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// ServeWs upgrades HTTP requests to WebSocket connections for private chat
func (h *Hub) ServeWs(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]any{
			"status":  "error",
			"message": "Method not allowed. Only GET is supported.",
		})
		http.Redirect(w, req, "/unauthorized", http.StatusSeeOther)
		return
	}

	accept := req.Header.Get("Accept")
	if containsHTML(accept) {
		// If browser wants HTML, redirect to unauthorized
		http.Redirect(w, req, "/unauthorized", http.StatusSeeOther)
		return
	}
	roomName := req.URL.Query().Get("room")
	user1 := req.URL.Query().Get("user1")
	user2 := req.URL.Query().Get("user2")
	currentUser := req.URL.Query().Get("username")

	if currentUser == "" {
		currentUser = "Anonymous"
	}

	log.Printf("WebSocket connection request - Room: %s, User1: %s, User2: %s, Current: %s",
		roomName, user1, user2, currentUser)

	// Get current user ID for database operations
	var currentUserID int
	if sessionCookie, err := req.Cookie("session_token"); err == nil {
		if userID, hasSession, err := db.SelectUserSession(sessionCookie.Value); err == nil && hasSession {
			currentUserID = userID
			log.Printf("Current user ID from session: %d", currentUserID)
		}
	}

	// Case 1: Public room
	if roomName != "" {
		socket, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			log.Println("Upgrade error:", err)
			return
		}

		r := h.GetRoom(roomName)
		user := &user{
			name:    currentUser,
			socket:  socket,
			recieve: make(chan []byte, messageBuffersize),
			room:    r,
			userID:  currentUserID,
		}
		r.join <- user

		// Add to global users but don't manage online status via WebSocket
		h.globalUsers[user.name] = user
		if currentUser != "" {
			auth.AddUserConnection(currentUser)
		}

		defer func() {
			r.leave <- user
			delete(h.globalUsers, user.name)
			if currentUser != "" {
				auth.RemoveUserConnection(currentUser)
			}
		}()
		go user.write()
		user.read()
		return
	}

	// Case 2: Private chat
	if user1 == "" || user2 == "" {
		http.Error(w, "Must specify room or both user1/user2", http.StatusBadRequest)
		return
	}

	if currentUser != user1 && currentUser != user2 {
		http.Error(w, "You can only join your own private chats", http.StatusForbidden)
		return
	}

	roomName = CreatePrivateRoomName(user1, user2)
	log.Printf("Creating/joining private room: %s", roomName)

	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	r := h.GetRoom(roomName)
	user := &user{
		name:    currentUser,
		socket:  socket,
		recieve: make(chan []byte, messageBuffersize),
		room:    r,
		userID:  currentUserID,
	}

	log.Printf("DEBUG: Created user struct: name=%s, userID=%d, room=%s", user.name, user.userID, r.name)

	r.join <- user
	log.Printf("DEBUG: User sent to join channel")

	// Add to global hub but track online status through connection counts
	h.globalUsers[user.name] = user
	if currentUser != "" {
		auth.AddUserConnection(currentUser)
	}
	log.Printf("DEBUG: User added to global users")

	defer func() {
		log.Printf("DEBUG: Defer cleanup starting for user %s", user.name)
		r.leave <- user
		delete(h.globalUsers, user.name)
		if currentUser != "" {
			auth.RemoveUserConnection(currentUser)
		}
		log.Printf("DEBUG: Defer cleanup complete for user %s", user.name)
	}()

	log.Printf("DEBUG: Starting write() goroutine for user %s", user.name)
	go user.write()
	log.Printf("DEBUG: Starting read() for user %s", user.name)
	user.read()
	log.Printf("DEBUG: read() exited for user %s", user.name)
}

// ServeNotifications handles WebSocket connections for global notifications
func (h *Hub) ServeNotifications(w http.ResponseWriter, req *http.Request) {
	username := req.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Username required", http.StatusBadRequest)
		return
	}

	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		return
	}

	user := &user{
		name:    username,
		socket:  socket,
		recieve: make(chan []byte, messageBuffersize),
	}

	h.notificationUsers[username] = user
	if username != "" {
		auth.AddUserConnection(username)
	}

	defer func() {
		delete(h.notificationUsers, username)
		if username != "" {
			auth.RemoveUserConnection(username)
		}
		socket.Close()
	}()

	go user.write()
	user.read()
}
