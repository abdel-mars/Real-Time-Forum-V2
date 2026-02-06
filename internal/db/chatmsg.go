package db

import (
	repo "forum/internal/repository"
	"fmt"
	"time"
)

type ChatMessage struct {
	ID         int    `json:"id"`
	SenderID   int    `json:"sender_id"`
	ReceiverID int    `json:"receiver_id"`
	Message    string `json:"message"`
	CreatedAt  string `json:"created_at"`
	IsRead     bool   `json:"is_read"`
	SenderName string `json:"sender_name"`
}

// SaveChatMessage saves a message to the database
func SaveChatMessage(senderID, receiverID int, message string) error {
	query := `INSERT INTO chat_messages (sender_id, receiver_id, message, created_at) 
			  VALUES (?, ?, ?, ?)`

	_, err := repo.DB.Exec(query, senderID, receiverID, message, time.Now().UnixMilli())
	return err
}

// GetChatMessages retrieves chat messages between two users with pagination
func GetChatMessages(userID1, userID2 int, limit, offset int) ([]ChatMessage, error) {
	query := `
		SELECT cm.id, cm.sender_id, cm.receiver_id, cm.message, cm.created_at, cm.is_read, u.username
		FROM chat_messages cm
		JOIN users u ON cm.sender_id = u.id
		WHERE (cm.sender_id = ? AND cm.receiver_id = ?) 
		   OR (cm.sender_id = ? AND cm.receiver_id = ?)
		ORDER BY cm.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := repo.DB.Query(query, userID1, userID2, userID2, userID1, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []ChatMessage
	for rows.Next() {
		var msg ChatMessage
		var createdAtMs int64 // created_at is stored as milliseconds
		err := rows.Scan(&msg.ID, &msg.SenderID, &msg.ReceiverID,
			&msg.Message, &createdAtMs, &msg.IsRead, &msg.SenderName)
		if err != nil {
			return nil, err
		}
		// Convert Unix milliseconds to ISO 8601 format WITH millisecond precision
		// Use RFC3339Nano so we preserve sub-second precision (milliseconds)
		msg.CreatedAt = time.UnixMilli(createdAtMs).UTC().Format(time.RFC3339Nano)
		messages = append(messages, msg)
	}

	// Reverse the slice to get chronological order (oldest first)
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// MarkMessagesAsRead marks messages as read
func MarkMessagesAsRead(senderID, receiverID int) error {
	query := `UPDATE chat_messages SET is_read = 1 
			  WHERE sender_id = ? AND receiver_id = ? AND is_read = 0`

	_, err := repo.DB.Exec(query, senderID, receiverID)
	return err
}

// GetUserByUsername gets user ID and basic info by username
func GetUserByUsername(username string) (int, string, error) {
	var userID int
	var email string
	query := `SELECT id, email FROM users WHERE username = ?`
	err := repo.DB.QueryRow(query, username).Scan(&userID, &email)
	if err != nil {
		return 0, "", err
	}
	return userID, email, nil
}

// GetUserIDByUsername gets just the user ID by username
func GetUserIDByUsername(username string) (int, error) {
	var userID int
	query := `SELECT id FROM users WHERE username = ?`
	err := repo.DB.QueryRow(query, username).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

// GetUnreadMessageCount gets the count of unread messages for a user
func GetUnreadMessageCount(userID int) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM chat_messages WHERE receiver_id = ? AND is_read = 0`
	err := repo.DB.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetRecentChatUsers gets users that the current user has recently chatted with
func GetRecentChatUsers(userID int, limit int) ([]map[string]interface{}, error) {
	query := `
		SELECT 
			u.id as user_id,
			u.username,
			MAX(cm.created_at) as last_message_time,
			SUM(CASE WHEN cm.receiver_id = ? AND cm.is_read = 0 AND cm.sender_id = u.id THEN 1 ELSE 0 END) as unread_count
		FROM chat_messages cm
		JOIN users u 
			ON (u.id = cm.sender_id AND cm.receiver_id = ?)
			OR (u.id = cm.receiver_id AND cm.sender_id = ?)
		WHERE cm.sender_id = ? OR cm.receiver_id = ?
		GROUP BY u.id, u.username
		ORDER BY last_message_time DESC
		LIMIT ?
	`

	rows, err := repo.DB.Query(query, userID, userID, userID, userID, userID, limit)
	if err != nil {
		fmt.Println("GetRecentChatUsers SQL error:", err) // log the SQL error
		return []map[string]interface{}{}, nil            // return empty list instead of crashing
	}

	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var chatUserID int
		var username string
		var lastMessageTime time.Time
		var unreadCount int

		err := rows.Scan(&chatUserID, &username, &lastMessageTime, &unreadCount)
		if err != nil {
			return nil, err
		}

		users = append(users, map[string]interface{}{
			"user_id":           chatUserID,
			"username":          username,
			"last_message_time": lastMessageTime,
			"unread_count":      unreadCount,
		})
	}

	return users, nil
}

func GetLastMessagesForUser(userID int) ([]map[string]interface{}, error) {
	query := `
        SELECT 
            u.username,
            MAX(cm.created_at) as last_message_time
        FROM chat_messages cm
        JOIN users u ON (
            (cm.sender_id = u.id AND cm.receiver_id = ?) OR 
            (cm.receiver_id = u.id AND cm.sender_id = ?)
        )
        WHERE u.id != ?
        GROUP BY u.id, u.username
        ORDER BY last_message_time DESC
    `

	rows, err := repo.DB.Query(query, userID, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []map[string]interface{}
	for rows.Next() {
		var username string
		var lastMessageTime time.Time

		err := rows.Scan(&username, &lastMessageTime)
		if err != nil {
			return nil, err
		}

		messages = append(messages, map[string]interface{}{
			"username":          username,
			"last_message_time": lastMessageTime,
		})
	}

	return messages, nil
}
