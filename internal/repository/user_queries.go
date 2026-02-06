package repository

const (
	// insert queries
	INSERT_NEW_SESSION             = `INSERT INTO sessions (user_id, session_token, expires_at) VALUES (?, ?, DATETIME('now', '+1 hour'))`
	INSERT_NEW_USER                = `INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)`
	INSERT_USERNAME_EMAIL_PASSHASH =  `
    INSERT INTO users (username, email, password_hash, first_name, last_name, age, gender)
    	VALUES (?, ?, ?, ?, ?, ?, ?)
	`


	// select queries
	SELECT_USER_BY_ID                        = `SELECT * FROM users WHERE id = ?`
	SELECT_USER_BY_SESSION_TOKEN             = `SELECT user_id FROM sessions WHERE session_token = ?`
	SELECT_USER_COUNT_BY_USERNAME_EMAIL      = `SELECT COUNT(*) FROM users WHERE username = ? OR email = ?`
	SELECT_USERID_PASSHASH_BY_USERNAME_EMAIL = `SELECT id,password_hash FROM users WHERE username = ? OR email = ?`
	SELECT_PASSHASH_BY_USERID                = `SELECT password_hash FROM users WHERE id = ?`
	CHECK_EMAIL_DUP                          = `SELECT COUNT(*) FROM users WHERE email = ?`
	CHECK_USERNAME_DUP                       = `SELECT COUNT(*) FROM users WHERE username = ?`
	SELECT_USERNAME_BY_ID                    = `SELECT username FROM users WHERE id = ?`
	SELECT_TIME                              = `SELECT created_at,updated_at FROM users WHERE id = ?`
	SELECT_ALL_USERNAMES                     = `SELECT username FROM users ORDER BY username`

	// update queries
	UPDATE_SESSION_EXPIRING_TIME = `UPDATE sessions SET expires_at = DATETIME('now', '+1 hour'), session_token = ? WHERE user_id = ?`
	RESET_USER_SESSION_TOKEN     = `UPDATE sessions SET session_token = NULL WHERE session_token = ?`
	UPDATE_PASS                  = `UPDATE users SET updated_at = DATETIME('now'), password_hash = ? WHERE id = ?`
	UPDATE_EMAIL                 = `UPDATE users SET updated_at = DATETIME('now') , email = ? WHERE id = ?`
	UPDATE_USER_NAME             = `UPDATE users SET updated_at = DATETIME('now') , username = ? WHERE id = ?`

	// delete queries
	DELETE_USER = `DELETE FROM users WHERE id = ?;
				UPDATE post_metadata SET post_count = (SELECT COUNT(*) FROM posts);`
)
