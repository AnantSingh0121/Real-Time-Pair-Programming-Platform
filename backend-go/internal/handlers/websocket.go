package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/anant/realtime-pair-programming/internal/db"
	"github.com/anant/realtime-pair-programming/internal/models"
	"github.com/anant/realtime-pair-programming/internal/services"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true 
	},
}

type WebSocketHandler struct {
	RoomManager *services.RoomManager
	DB          *db.DynamoDB
}

func NewWebSocketHandler(rm *services.RoomManager, database *db.DynamoDB) *WebSocketHandler {
	return &WebSocketHandler{
		RoomManager: rm,
		DB:          database,
	}
}

func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomId")
	userID := r.URL.Query().Get("userId")
	username := r.URL.Query().Get("username")

	if roomID == "" || userID == "" || username == "" {
		http.Error(w, "Missing roomId, userId, or username", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &services.Client{
		ConnID:   uuid.New().String(),
		UserID:   userID,
		Username: username,
		RoomID:   roomID,
		Conn:     conn,
		Send:     make(chan []byte, 256),
	}

	h.RoomManager.RegisterClient(client)

	go h.writePump(client)
	go h.readPump(client)
}

func (h *WebSocketHandler) readPump(client *services.Client) {
	defer func() {
		h.RoomManager.UnregisterClient(client)
		client.Conn.Close()
	}()

	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var wsMsg models.WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		h.handleMessage(client, &wsMsg)
	}
}

func (h *WebSocketHandler) writePump(client *services.Client) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *WebSocketHandler) handleMessage(client *services.Client, msg *models.WSMessage) {
	switch msg.Type {
	case "code_change":
		h.handleCodeChange(client, msg)
	case "chat":
		h.handleChat(client, msg)
	case "cursor":
		h.handleCursor(client, msg)
	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

func (h *WebSocketHandler) handleCodeChange(client *services.Client, msg *models.WSMessage) {
	payloadBytes, _ := json.Marshal(msg.Payload)
	var payload models.CodeChangePayload
	json.Unmarshal(payloadBytes, &payload)
	_, err := h.DB.Client.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName: aws.String(h.DB.CodeSyncTable),
		Key: map[string]types.AttributeValue{
			"roomId": &types.AttributeValueMemberS{Value: client.RoomID},
		},
		UpdateExpression: aws.String("SET code = :code, updatedAt = :now, #lang = :lang"),
		ExpressionAttributeNames: map[string]string{
			"#lang": "language",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":code": &types.AttributeValueMemberS{Value: payload.Code},
			":now":  &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
			":lang": &types.AttributeValueMemberS{Value: payload.Language},
		},
	})
	if err != nil {
		log.Printf("Error updating code: %v", err)
	}

	broadcastMsg, _ := json.Marshal(msg)
	h.RoomManager.BroadcastToRoom(client.RoomID, broadcastMsg, client.UserID)
}

func (h *WebSocketHandler) handleChat(client *services.Client, msg *models.WSMessage) {
	payloadBytes, _ := json.Marshal(msg.Payload)
	var payload models.ChatPayload
	json.Unmarshal(payloadBytes, &payload)

	message := models.Message{
		RoomID:    client.RoomID,
		MessageID: uuid.New().String(),
		UserID:    client.UserID,
		Username:  client.Username,
		Text:      payload.Text,
		Timestamp: time.Now(),
	}

	item, _ := attributevalue.MarshalMap(message)
	h.DB.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(h.DB.MessagesTable),
		Item:      item,
	})

	responseMsg := models.WSMessage{
		Type:    "chat",
		Payload: message,
	}
	broadcastData, _ := json.Marshal(responseMsg)
	h.RoomManager.BroadcastToRoom(client.RoomID, broadcastData, "")
}

func (h *WebSocketHandler) handleCursor(client *services.Client, msg *models.WSMessage) {
	broadcastMsg, _ := json.Marshal(msg)
	h.RoomManager.BroadcastToRoom(client.RoomID, broadcastMsg, client.UserID)
}
