package db

import (
	repo "forum/internal/repository"
)

// GetAllUsernames returns all usernames ordered alphabetically
func GetAllUsernames() ([]string, error) {
	rows, err := repo.DB.Query(repo.SELECT_ALL_USERNAMES)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []string
	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err != nil {
			return nil, err
		}
		users = append(users, username)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}
