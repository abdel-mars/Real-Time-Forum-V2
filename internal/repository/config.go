package repository

import (
	"database/sql"
	"html/template"
	"regexp"
)

var (
	EmailExp 		*regexp.Regexp
	UsernameExp 	*regexp.Regexp
	GLOBAL_TEMPLATE *template.Template
	DB 				*sql.DB
)

type User struct {
	Id       	  int
	Username 	  string
	Email		  string
	Password_hash string
	Created_at 	  string
	Updated_at	  string
}

type Post struct {
	Id       	  	 int
	Title		  	 string
	PublisherId   	 int
	Content 	  	 string
	Publisher 	  	 string
	Initial 	  	 string
	Catigories 	  	 []string
	Created_at 	  	 string
	Updated_at	  	 string
	Likes		  	 int
	Dislikes	  	 int
	CommentsCount  	 int
	IsEdited	  	 bool
	IsLikedByUser 	 bool
	IsDislikedByUser bool
	HasCategories 	 bool
	Owned 		  	 bool
}

type PageData struct {
	Posts []Post
}

type Comment struct {
	Username string        `json:"username"`
	Initial  string        `json:"initial"`
	Content  template.HTML `json:"content"` // keep template.HTML

	DateCreated string `json:"dateCreated"`
	PostId      int    `json:"postId"`
	CommentId   int    `json:"commentId"`

	IsCommentLikedByUser    bool `json:"isCommentLikedByUser"`
	IsCommentDislikedByUser bool `json:"isCommentDislikedByUser"`
	CommentLikes            int  `json:"commentLikes"`
	CommentDislikes         int  `json:"commentDislikes"`
}

// fit with json data who i will send to front
type PostResponse struct {
	Success       bool      `json:"success"`
	Post          Post      `json:"post"`
	Comments      []Comment `json:"comments"`
	CurrentPage   int       `json:"currentPage"`
	HasPrev       bool      `json:"hasPrev"`
	PrevPage      string    `json:"prevPage,omitempty"`
	HasNext       bool      `json:"hasNext"`
	NextPage      string    `json:"nextPage,omitempty"`
	Authenticated bool      `json:"authenticated"`
	Username      string    `json:"username,omitempty"`
	Initial       string    `json:"initial,omitempty"`
}

type contextKey string

const (
	INTERNAL_SERVER_ERROR_LOG_PATH = "./logs/internal_errors.log"

	DATABASE_NAME = "sqlite3"

	// database paths
	DATABASE_LOCATION        = "./database/forum.db"
	DATABASE_SCHEMA_LOCATION = "./database/schema.sql"

	// error messages
	FAILED_OPEN_DATABES     = "failed to open the database: %v"
	FAILED_CREAT_TABELS     = "failed to create tables: %v"
	FAILED_CLOSING_DATABASE = "closing database: %v"

	// templates paths
	TEMPLATE_PATHS = "./templates/*.html"

	// server helpers
	PORT               = ":8081"
	SERVER_RUN_MESSAGE = "\033[2mServer running on http://localhost:8081\033[0m"

	// context keys
	USER_ID_KEY contextKey = "userId"
	ERROR_CASE  contextKey = "error_case"
	USER_NAME   contextKey = "userName"

	PAGE_POSTS_QUANTITY   = 10
	PAGE_COMMENT_QUANTITY = 10
	DAY_POST_LIMIT        = 20
	DAY_COMMENTS_LIMIT    = 50
)

// IT major fields
var IT_MAJOR_FIELDS = map[string]bool{
	"Software Engineering": true,
	"Cybersecurity": true,
	"DevOps and SRE":   true,
	"Database Systems": true,
}

// users limitations
const (
	// Email limitations
	EMAIL_MIN_LEN = 6   // a@b.co is a valid short email
	EMAIL_MAX_LEN = 254 // RFC 5321 max length of an email address

	// Username limitations
	USERNAME_MIN_LEN = 3
	USERNAME_MAX_LEN = 32 // Long enough, yet avoids abuse or awkward UI

	// Password limitations
	PASSWORD_MIN_LEN = 8  // Minimum for secure password
	PASSWORD_MAX_LEN = 72 // Bcrypt max input length

	// Post Title limitations
	TITLE_MIN_LEN = 10
	TITLE_MAX_LEN = 70 // Enough for concise titles; avoids clutter

	// Post Content limitations
	POST_MIN_LEN = 1
	POST_MAX_LEN = 10_000 // Long enough for article-style posts

	// Comment limitations
	COMMENT_MIN_LEN = 1
	COMMENT_MAX_LEN = 1_000 // Reasonable upper bound for a comment
)
