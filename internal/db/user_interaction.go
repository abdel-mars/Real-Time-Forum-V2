package db

import (
	repo "forum/internal/repository"
	"database/sql"
	"errors"
	"log"
	"time"

)

func SelectUserSession(session_id string) (int, bool, error) {
	var userId int

	err := repo.DB.QueryRow(repo.SELECT_USER_BY_SESSION_TOKEN, session_id).Scan(&userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return userId, false, nil // not logged in
		}
		return userId, false, err // err in database
	}
	return userId, true, nil // logged in
}

func UpdateUserSession(id int, token string) error {
	tx, err := repo.DB.Begin()
	if err != nil {
		log.Println("Failed to start transaction:", err)
		return err
	}

	res, err := tx.Exec(repo.UPDATE_SESSION_EXPIRING_TIME, token, id)
	if err != nil {
		log.Println("Update error:", err)
		tx.Rollback()
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		log.Println("Error getting affected rows:", err)
		tx.Rollback()
		return err
	}

	if count > 0 {
		return tx.Commit() // if update succed
	}

	_, err = tx.Exec(repo.INSERT_NEW_SESSION, id, token)
	if err != nil {
		log.Println("Insert error:", err)
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func ResetUserSession(session_id string) (bool, error) {
	_, err := repo.DB.Exec(repo.RESET_USER_SESSION_TOKEN, session_id)
	if err != nil {
		return false, err
	}
	return true, nil
}

func AlreadyExists(username, email string) (bool, error) {
	var count int
	err := repo.DB.QueryRow(repo.SELECT_USER_COUNT_BY_USERNAME_EMAIL, username, email).Scan(&count)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}

func GetUserHashByUsername(username string) (int, string, error) {
	var hash string
	var id int
	err := repo.DB.QueryRow(repo.SELECT_USERID_PASSHASH_BY_USERNAME_EMAIL, username, username).Scan(&id, &hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return id, hash, nil
		}
		return id, hash, err
	}
	return id, hash, nil
}

func GetUserInfo(userId int) (repo.User, error) {
	var user repo.User

	err := repo.DB.QueryRow(repo.SELECT_USER_BY_ID, userId).Scan(&user.Id, &user.Username,
		&user.Email, &user.Password_hash, &user.Created_at, &user.Updated_at)
	if err != nil {
		return user, err
	}
	return user, nil
}

func AddNewUser(username, email, hashedPass, firstName, lastName, gender string, age int) error {
	_, err := repo.DB.Exec(
		repo.INSERT_USERNAME_EMAIL_PASSHASH,
		username, email, hashedPass, firstName, lastName, age, gender,
	)
	return err
}

func GetUserHashById(id int) (string, error) {
	var hash string
	err := repo.DB.QueryRow(repo.SELECT_PASSHASH_BY_USERID, id).Scan(&hash)
	if err != nil {
		return hash, err
	}
	return hash, nil
}

func DupplicatedUsername(username string) (bool, error) {
	var count int
	err := repo.DB.QueryRow(repo.CHECK_USERNAME_DUP, username).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func UpdateUsernmae(id int, username string) error {
	res, err := repo.DB.Exec(repo.UPDATE_USER_NAME, username, id)
	if err != nil {
		return err
	}
	_, err = res.RowsAffected()
	return err
}

func DupplicatedEmail(email string) (bool, error) {
	var count int
	err := repo.DB.QueryRow(repo.CHECK_EMAIL_DUP, email).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func UpdateEmail(id int, email string) error {
	res, err := repo.DB.Exec(repo.UPDATE_EMAIL, email, id)
	if err != nil {
		return err
	}
	_, err = res.RowsAffected()
	return err
}

func UpdatePassword(id int, password string) error {
	res, err := repo.DB.Exec(repo.UPDATE_PASS, password, id)
	if err != nil {
		return err
	}
	_, err = res.RowsAffected()
	return err
}

func DeleteUser(userId int) error {
	_, err := repo.DB.Exec(repo.DELETE_USER, userId, userId)
	if err != nil {
		return err
	}
	err = UpdateCatCount()
	if err != nil {
		return err
	}
	return nil
}

func GetUserNameById(userId int) (string, error) {
	var userName string

	err := repo.DB.QueryRow(repo.SELECT_USERNAME_BY_ID, userId).Scan(&userName)
	if err != nil {
		return userName, err
	}
	return userName, nil
}

func IsUpdateAllowed(userId int) (bool, error) {
	now := time.Now().UTC()
	var created, updated time.Time // uint int *time.Location == nil utc
	err := repo.DB.QueryRow(repo.SELECT_TIME, userId).Scan(&created, &updated)
	if err != nil {
		return false, err
	}
	diff := now.Sub(updated)
	if diff < (time.Hour*72) && created != updated {
		return false, nil
	}
	return true, nil
}

// SetUserOnline sets the user's online status to true
func SetUserOnline(userID int) error {
	query := `UPDATE users SET online = 1 WHERE id = ?`
	_, err := repo.DB.Exec(query, userID)
	return err
}

// SetUserOffline sets the user's online status to false
func SetUserOffline(userID int) error {
	query := `UPDATE users SET online = 0 WHERE id = ?`
	_, err := repo.DB.Exec(query, userID)
	return err
}

// GetOnlineUsers retrieves all online users
func GetOnlineUsers() ([]string, error) {
	query := `SELECT username FROM users WHERE online = 1`
	rows, err := repo.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usernames []string
	for rows.Next() {
		var username string
		err := rows.Scan(&username)
		if err != nil {
			return nil, err
		}
		usernames = append(usernames, username)
	}
	return usernames, nil
}

// ResetAllUsersOffline sets all users to offline status
func ResetAllUsersOffline() error {
	query := `UPDATE users SET online = 0`
	_, err := repo.DB.Exec(query)
	return err
}
