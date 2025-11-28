package models

import "time"

type User struct {
	UserID         string    `json:"userId" dynamodbav:"userId"`
	Username       string    `json:"username" dynamodbav:"username"`
	Email          string    `json:"email" dynamodbav:"email"`
	HashedPassword string    `json:"-" dynamodbav:"hashedPassword"`
	CreatedAt      time.Time `json:"createdAt" dynamodbav:"createdAt"`
	LastSeen       time.Time `json:"lastSeen" dynamodbav:"lastSeen"`
}

type Room struct {
	RoomID    string    `json:"roomId" dynamodbav:"roomId"`
	Name      string    `json:"name" dynamodbav:"name"`
	CreatedBy string    `json:"createdBy" dynamodbav:"createdBy"`
	Users     []string  `json:"users" dynamodbav:"users"`
	CreatedAt time.Time `json:"createdAt" dynamodbav:"createdAt"`
}

type Message struct {
	RoomID    string    `json:"roomId" dynamodbav:"roomId"`
	MessageID string    `json:"messageId" dynamodbav:"messageId"`
	UserID    string    `json:"userId" dynamodbav:"userId"`
	Username  string    `json:"username" dynamodbav:"username"`
	Text      string    `json:"text" dynamodbav:"text"`
	Timestamp time.Time `json:"timestamp" dynamodbav:"timestamp"`
}

type CodeSync struct {
	RoomID    string    `json:"roomId" dynamodbav:"roomId"`
	Code      string    `json:"code" dynamodbav:"code"`
	Language  string    `json:"language" dynamodbav:"language"`
	UpdatedAt time.Time `json:"updatedAt" dynamodbav:"updatedAt"`
}

type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type CodeChangePayload struct {
	RoomID   string `json:"roomId"`
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Code     string `json:"code"`
	Language string `json:"language"`
	Changes  string `json:"changes"` 
}

type ChatPayload struct {
	RoomID    string    `json:"roomId"`
	UserID    string    `json:"userId"`
	Username  string    `json:"username"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}

type CursorPayload struct {
	RoomID   string `json:"roomId"`
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Position struct {
		LineNumber int `json:"lineNumber"`
		Column     int `json:"column"`
	} `json:"position"`
}

type UserPresence struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Status   string `json:"status"` 
}

type SignupRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token  string `json:"token"`
	UserID string `json:"userId"`
	User   User   `json:"user"`
}

type CreateRoomRequest struct {
	Name string `json:"name"`
}

type JoinRoomRequest struct {
	RoomID string `json:"roomId"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
