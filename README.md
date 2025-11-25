# Real-Time Chat Application in Go

A real-time chat application built from the ground up using Go (Golang). The primary goal of this project is to gain a deep, fundamental understanding of Go's standard library, concurrency patterns, and clean architecture principles.

---

## 🌟 Project Goals

- **Deep Dive into Go:** Understand Go's core features by building a substantial project without the abstractions of a framework.
- **Clean Architecture:** Implement a scalable and maintainable project structure by separating concerns (domain, repository, service, handler).
- **Pure WebSocket Implementation:** Handle real-time communication using Go's standard library and the Gorilla WebSocket library, avoiding higher-level abstractions like Socket.io.
- **Database Fundamentals:** Interact with a PostgreSQL database using the standard `database/sql` package to master SQL operations and transactions in Go.
- **Professional Practices:** Adhere to good software engineering habits, including structured commits, clear documentation, and configuration management.

## ✨ Features (Planned)

- [ ] User Authentication (Registration & Login)
- [ ] Contact Management (Adding contacts who are registered or unregistered)
- [ ] One-on-One Private Messaging
- [ ] Group Chats with Multiple Participants
- [ ] Real-time Message Status (Sent, Delivered, Read)
- [ ] Media Sharing (Images, Files)

## 🛠️ Tech Stack

- **Language:** Go (Golang)
- **Database:** PostgreSQL
- **Real-time Communication:** WebSockets (via `gorilla/websocket`)
- **Password Hashing:** `golang.org/x/crypto/bcrypt`

## 🚀 Getting Started

### Prerequisites

- [Go](https://go.dev/doc/install) (version 1.25.3 or higher)
- [PostgreSQL](https://www.postgresql.org/download/)
- A running PostgreSQL instance.

### Installation & Running

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/ErisSusanto19/chat-app-v2-backend
    cd chat-app-v2-backend
    ```

2.  **Create a configuration file:**
    Copy the example environment file and update it with your database credentials.
    ```bash
    cp .env.example .env
    ```
    
    Update the `DATABASE_URL` in the `.env` file:
    ```
    DATABASE_URL="postgres://user:password@localhost:5432/chat_go_db?sslmode=disable"
    ```

3.  **Install dependencies:**
    ```bash
    go mod tidy
    ```

4.  **Run the application:**
    ```bash
    go run ./cmd/server/main.go
    ```

The server will start on the configured port.

---
_This project is a work in progress, built for learning and demonstration purposes._