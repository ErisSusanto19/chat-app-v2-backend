FROM golang:1.24.0-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /chat-app ./cmd/server

FROM alpine:latest

COPY --from=builder /chat-app /chat-app

RUN apk --no-cache add ca-certificates

EXPOSE 8080

CMD ["/chat-app"]