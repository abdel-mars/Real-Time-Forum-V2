package service

import (
	"fmt"
	"log"
	"net/http"

	auth "forum/internal/auth"
	db "forum/internal/db"
	handler "forum/internal/handler"
	middleware "forum/internal/middleware"
	repo "forum/internal/repository"
	utils "forum/internal/utils"

	_ "github.com/mattn/go-sqlite3"
)

func InitDependencies() {
	db.InitDB(repo.DATABASE_LOCATION)
	utils.InitRegex()
	// reset all users offline when server start
	db.ResetAllUsersOffline()
	auth.LoadOnlineUsersFromDB()
}

func forumMux() *http.ServeMux {
	forumux := http.NewServeMux()

	// Create the WebSocket hub for managing connections
	hub := handler.NewHub()
	// Set the global hub in the auth package to avoid import cycles
	auth.GlobalHub = hub

	// API endpoints for user and application state
	forumux.HandleFunc("/api/me", middleware.InjectUser(handler.FirststateHandler))
	forumux.HandleFunc("/api/posts", middleware.InjectUser(handler.PostsHandlerApi))
	// API endpoint for retrieving online users for chat functionality
	forumux.HandleFunc("/api/online-users", middleware.InjectUser(handler.OnlineUsersHandler))
	forumux.HandleFunc("/api/all-users", middleware.InjectUser(handler.AllUsersHandler))
	forumux.HandleFunc("/api/chat-messages", handler.GetChatMessagesHandler)
	forumux.HandleFunc("/api/user-by-username", handler.GetUserByUsernameHandler)
	forumux.HandleFunc("/api/recent-chats", handler.GetRecentChatsHandler)
	forumux.HandleFunc("/api/unread-count", handler.GetUnreadCountHandler)
	forumux.HandleFunc("/api/last-messages", middleware.InjectUser(handler.GetLastMessagesHandler))

	// Authentication routes
	forumux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			auth.SubmitLogin(w, r)
		} else {
			handler.RootHandler(w, r)
		}
	})
	forumux.HandleFunc("/logout", auth.LogoutHandler)
	forumux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			auth.SubmitRegister(w, r)
		} else {
			handler.RootHandler(w, r)
		}
	})

	// Post-related routes
	forumux.HandleFunc("/post", middleware.InjectUser(handler.PostHandler))
	forumux.HandleFunc("/newPost", middleware.AuthMidleware(handler.PostPostHandler))

	// Like and dislike functionality
	forumux.HandleFunc("/like", middleware.AuthMidleware(handler.LikeHandler))
	forumux.HandleFunc("/dislike", middleware.AuthMidleware(handler.DislikeHandler))

	// Comment functionality
	forumux.HandleFunc("/comment", middleware.AuthMidleware(handler.CommentHandler))

	// Static file serving: serve ui assets directly and keep legacy /static/ handler
	forumux.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./ui/js"))))
	forumux.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./ui/css"))))
	forumux.Handle("/svg/", http.StripPrefix("/svg/", http.FileServer(http.Dir("./ui/svg"))))
	forumux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./ui/assets"))))
	// keep legacy /static/ route for compatibility
	forumux.HandleFunc("/static/", handler.StaticHandler)

	// WebSocket endpoints
	// WebSocket endpoint for private chat (requires authentication)
	forumux.HandleFunc("/room", hub.ServeWs)
	// WebSocket endpoint for global notifications
	forumux.HandleFunc("/notifications", hub.ServeNotifications)

	// Root handler for the main page
	forumux.HandleFunc("/", middleware.InjectUser(handler.RootHandler))

	return forumux
}

func StartServer() {
	server := &http.Server{
		Addr:    repo.PORT,
		Handler: middleware.RateLimiterMiddleware(forumMux(), 15, 30), // Rate limiter: 15 requests per 30 seconds
	}

	fmt.Println(repo.SERVER_RUN_MESSAGE)

	// Start the server in a goroutine to allow for signal handling
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			db.CloseDB()
			log.Fatalf("server error: %v", err)
		}
	}()

	// Handle OS signals for graceful shutdown
	HandleSignals(server)
}
