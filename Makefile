.PHONY: run run-backend run-frontend help

run: run-backend run-frontend

run-backend:
	@echo "Starting backend server..."
	go run cmd/forum/main.go

run-frontend:
	@echo "Starting frontend server..."
	npm run dev

	help:
	@echo "Available commands:"
	@echo "  run           - Start both backend and frontend servers"
	@echo "  run-backend   - Start only the backend server"
	@echo "  run-frontend  - Start only the frontend server"
	@echo "  help          - Show this help message"