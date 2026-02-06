package handler

import (
	db "forum/internal/db"
	"encoding/json"
	"log"
	"net/http"
	"github.com/gorilla/websocket"
)

type user struct {
	name    string
	socket  *websocket.Conn
	recieve chan []byte
	room    *room
	userID  int
}

func (c *user) read() {
	log.Printf("DEBUG read(): Starting for user %s", c.name)
	defer func() {
		log.Printf("DEBUG read(): Defer cleanup for user %s", c.name)
		if r := recover(); r != nil {
			log.Printf("PANIC in read() for user %s: %v", c.name, r)
		}
		if c.room != nil {
			log.Printf("DEBUG read(): Sending leave for user %s", c.name)
			c.room.leave <- c
		}
		if c.socket != nil {
			log.Printf("DEBUG read(): Closing socket for user %s", c.name)
			c.socket.Close()
		}
		log.Printf("DEBUG read(): Defer cleanup complete for user %s", c.name)
	}()

	for {
		log.Printf("DEBUG read(): Waiting for message for user %s", c.name)
		_, msg, err := c.socket.ReadMessage()
		if err != nil {
			log.Printf("DEBUG read(): Error reading for user %s: %v", c.name, err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for user %s: %v", c.name, err)
			} else {
				log.Printf("User %s disconnected normally", c.name)
			}
			return
		}

		log.Printf(" Received message from %s: %s", c.name, string(msg))

		if c.room != nil {
			// Add sender ID to message before forwarding
			var msgData map[string]interface{}
			if err := json.Unmarshal(msg, &msgData); err == nil {
				msgData["sender_id"] = c.userID
				msgData["created_at"] = "" // Will be set by database
				if modifiedMsg, err := json.Marshal(msgData); err == nil {
					select {
					case c.room.forward <- modifiedMsg:
						log.Printf(" Message forwarded to room with sender_id %d", c.userID)
					default:
						log.Printf(" Room channel full, message dropped")
					}
				} else {
					log.Printf(" Error marshaling modified message: %v", err)
				}
			} else {
				// Fallback: forward original message
				select {
				case c.room.forward <- msg:
					log.Printf(" Message forwarded to room (fallback)")
				default:
					log.Printf(" Room channel full, message dropped")
				}
			}
		}
	}
}

func (c *user) write() {
	log.Printf("DEBUG write(): Starting for user %s", c.name)
	defer func() {
		log.Printf("DEBUG write(): Defer cleanup for user %s", c.name)
		if r := recover(); r != nil {
			log.Printf("PANIC in write() for user %s: %v", c.name, r)
		}
		if c.socket != nil {
			log.Printf("DEBUG write(): Closing socket for user %s", c.name)
			c.socket.Close()
		}
		log.Printf("DEBUG write(): Defer cleanup complete for user %s", c.name)
	}()

	for msg := range c.recieve {
		log.Printf("DEBUG write(): Got message for user %s, message size: %d", c.name, len(msg))
		if c.socket == nil {
			log.Printf("Socket is nil for user %s", c.name)
			return
		}

		log.Printf("DEBUG write(): Writing message for user %s", c.name)
		err := c.socket.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Printf("Write error for user %s: %v", c.name, err)
			return
		}

		log.Printf("Message sent to %s: %s", c.name, string(msg))
	}
	log.Printf("DEBUG write(): Channel closed for user %s", c.name)
}

func GetLastMessagesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]any{
			"status":  "error",
			"message": "Method not allowed. Only GET is supported.",
		})
		http.Redirect(w, r, "/unauthorized", http.StatusSeeOther)
		return
	}

	accept := r.Header.Get("Accept")
	if containsHTML(accept) {
		// If browser wants HTML, redirect to unauthorized
		http.Redirect(w, r, "/unauthorized", http.StatusSeeOther)
		return
	}
	// Get current user from session (NOT from context)
	sessionCookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	currentUserID, hasSession, err := db.SelectUserSession(sessionCookie.Value)
	if err != nil || !hasSession {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get last messages for sorting
	lastMessages, err := db.GetLastMessagesForUser(currentUserID)
	if err != nil {
		log.Printf("Error getting last messages: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"last_messages": lastMessages,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
