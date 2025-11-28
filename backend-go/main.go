package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/anant/realtime-pair-programming/internal/auth"
	"github.com/anant/realtime-pair-programming/internal/db"
	"github.com/anant/realtime-pair-programming/internal/handlers"
	"github.com/anant/realtime-pair-programming/internal/services"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Printf("Warning: .env file not found, using system environment variables")
	}
	database, err := db.NewDynamoDB()
	if err != nil {
		log.Fatalf("Failed to initialize DynamoDB: %v", err)
	}
	if err := database.EnsureTablesExist(context.TODO()); err != nil {
		log.Fatalf("Failed to ensure tables exist: %v", err)
	}
	roomManager := services.NewRoomManager()
	go roomManager.Run()
	authHandler := handlers.NewAuthHandler(database)
	roomHandler := handlers.NewRoomHandler(database)
	wsHandler := handlers.NewWebSocketHandler(roomManager, database)
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Post("/api/auth/signup", authHandler.Signup)
	r.Post("/api/auth/login", authHandler.Login)
	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware)
		r.Get("/api/rooms", roomHandler.GetRooms)
		r.Post("/api/rooms", roomHandler.CreateRoom)
		r.Get("/api/rooms/{roomId}", roomHandler.GetRoom)
		r.Post("/api/rooms/{roomId}/join", roomHandler.JoinRoom)
	})
	r.Get("/ws/{roomId}", wsHandler.HandleWebSocket)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	port := os.Getenv("GO_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Golang backend server starting on port %s", port)
	log.Printf("WebSocket endpoint: ws://localhost:%s/ws/{roomId}", port)
	log.Printf("REST API: http://localhost:%s/api", port)

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
