package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/ErisSusanto19/chat-app-v2-backend/internal/config"
	"github.com/ErisSusanto19/chat-app-v2-backend/internal/handler"
	"github.com/ErisSusanto19/chat-app-v2-backend/internal/repository"
	"github.com/ErisSusanto19/chat-app-v2-backend/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	ws "github.com/ErisSusanto19/chat-app-v2-backend/internal/websocket"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {

	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("could not load config: %v", err)
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("could not ping database: %v", err)
	}

	fmt.Println("Successfully connected to the database!")
	fmt.Printf("Starting server on port %s\n", cfg.ServerPort)

	userRepo := repository.NewPostgresUserRepository(db)
	convRepo := repository.NewPostgresConversationRepository(db)
	msgRepo := repository.NewPostgresMessageRepository(db)
	contactRepo := repository.NewPostgresContactRepository(db)

	hub := ws.NewHub(nil)
	go hub.Run()

	authService := service.NewAuthService(userRepo, cfg.JWTSecretKey)
	chatService := service.NewChatService(msgRepo, convRepo, hub)
	contactService := service.NewContactService(contactRepo, userRepo)

	hub.SetChatService(chatService)

	authHandler := handler.NewAuthHandler(authService)
	wsHandler := handler.NewWebsocketHandler(hub)
	convHandler := handler.NewConversationHandler(chatService)
	contactHandler := handler.NewContactHandler(contactService)

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Route("/api/v1", func(r chi.Router) {

		r.Group(func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
		})

		r.Group(func(r chi.Router) {
			r.Use(handler.AuthMiddleware(cfg.JWTSecretKey))

			r.Get("/me", authHandler.Me)
			r.Put("/me", authHandler.UpdateProfile)
			r.Put("/me/password", authHandler.ChangePassword)

			r.Get("/ws", wsHandler.ServeWs)
			r.Post("/conversations", convHandler.StartPrivateConversation)
			r.Get("/conversations", convHandler.GetConversations)
			r.Get("/conversations/{conversationID}/messages", convHandler.GetMessages)

			r.Post("/contacts", contactHandler.AddContact)
			r.Get("/contacts", contactHandler.GetContacts)
			r.Put("/contacts/{contactID}", contactHandler.UpdateContact)
			r.Delete("/contacts/{contactID}", contactHandler.DeleteContact)
			r.Get("/contacts/{contactID}", contactHandler.GetContactDetail)

			r.Post("/conversations/group", convHandler.CreateGroup)
			r.Post("/conversations/{conversationID}/participants", convHandler.AddParticipants)
			r.Delete("/conversations/{conversationID}/participants/{userID}", convHandler.RemoveParticipant)
			r.Delete("/conversations/{conversationID}/leave", convHandler.LeaveGroup)
		})
	})

	log.Printf("Starting server on port %s", cfg.ServerPort)
	if err := http.ListenAndServe(cfg.ServerPort, router); err != nil {
		log.Fatalf("could not start server: %v", err)
	}

}
