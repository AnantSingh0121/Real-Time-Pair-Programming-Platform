package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/anant/realtime-pair-programming/internal/auth"
	"github.com/anant/realtime-pair-programming/internal/db"
	"github.com/anant/realtime-pair-programming/internal/models"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type RoomHandler struct {
	DB *db.DynamoDB
}

func NewRoomHandler(database *db.DynamoDB) *RoomHandler {
	return &RoomHandler{DB: database}
}

func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(string)
	username := r.Context().Value(auth.UsernameKey).(string)

	var req models.CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		req.Name = "Untitled Room"
	}

	room := models.Room{
		RoomID:    uuid.New().String(),
		Name:      req.Name,
		CreatedBy: userID,
		Users:     []string{userID},
		CreatedAt: time.Now(),
	}

	item, err := attributevalue.MarshalMap(room)
	if err != nil {
		http.Error(w, "Error creating room", http.StatusInternalServerError)
		return
	}

	_, err = h.DB.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(h.DB.RoomsTable),
		Item:      item,
	})
	if err != nil {
		http.Error(w, "Error saving room", http.StatusInternalServerError)
		return
	}

	codeSync := models.CodeSync{
		RoomID:    room.RoomID,
		Code:      "// Welcome to the pair programming session!\n// Start coding here...\n",
		Language:  "javascript",
		UpdatedAt: time.Now(),
	}

	codeSyncItem, _ := attributevalue.MarshalMap(codeSync)
	h.DB.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(h.DB.CodeSyncTable),
		Item:      codeSyncItem,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"room":    room,
		"message": "Room created successfully by " + username,
	})
}

func (h *RoomHandler) GetRooms(w http.ResponseWriter, r *http.Request) {
	result, err := h.DB.Client.Scan(context.TODO(), &dynamodb.ScanInput{
		TableName: aws.String(h.DB.RoomsTable),
	})
	if err != nil {
		http.Error(w, "Error fetching rooms", http.StatusInternalServerError)
		return
	}

	var rooms []models.Room
	err = attributevalue.UnmarshalListOfMaps(result.Items, &rooms)
	if err != nil {
		http.Error(w, "Error processing rooms", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rooms)
}

func (h *RoomHandler) GetRoom(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomId")

	result, err := h.DB.Client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(h.DB.RoomsTable),
		Key: map[string]types.AttributeValue{
			"roomId": &types.AttributeValueMemberS{Value: roomID},
		},
	})
	if err != nil {
		http.Error(w, "Error fetching room", http.StatusInternalServerError)
		return
	}
	if result.Item == nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	var room models.Room
	err = attributevalue.UnmarshalMap(result.Item, &room)
	if err != nil {
		http.Error(w, "Error processing room", http.StatusInternalServerError)
		return
	}

	codeSyncResult, _ := h.DB.Client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(h.DB.CodeSyncTable),
		Key: map[string]types.AttributeValue{
			"roomId": &types.AttributeValueMemberS{Value: roomID},
		},
	})

	var codeSync models.CodeSync
	if codeSyncResult.Item != nil {
		attributevalue.UnmarshalMap(codeSyncResult.Item, &codeSync)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"room":     room,
		"codeSync": codeSync,
	})
}

func (h *RoomHandler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomId")
	userID := r.Context().Value(auth.UserIDKey).(string)

	log.Printf("JoinRoom called: roomID=%s, userID=%s", roomID, userID)

	result, err := h.DB.Client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(h.DB.RoomsTable),
		Key: map[string]types.AttributeValue{
			"roomId": &types.AttributeValueMemberS{Value: roomID},
		},
	})
	if err != nil {
		log.Printf("Error fetching room: %v", err)
		http.Error(w, "Error fetching room", http.StatusInternalServerError)
		return
	}
	if result.Item == nil {
		log.Printf("Room not found: %s", roomID)
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	var room models.Room
	err = attributevalue.UnmarshalMap(result.Item, &room)
	if err != nil {
		log.Printf("Error unmarshaling room: %v", err)
		http.Error(w, "Error processing room", http.StatusInternalServerError)
		return
	}

	for _, uid := range room.Users {
		if uid == userID {
			log.Printf("User %s already in room %s", userID, roomID)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "Already in room",
				"room":    room,
			})
			return
		}
	}

	_, err = h.DB.Client.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName: aws.String(h.DB.RoomsTable),
		Key: map[string]types.AttributeValue{
			"roomId": &types.AttributeValueMemberS{Value: roomID},
		},
		UpdateExpression: aws.String("SET #users = list_append(if_not_exists(#users, :empty_list), :user)"),
		ExpressionAttributeNames: map[string]string{
			"#users": "users",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":user":       &types.AttributeValueMemberL{Value: []types.AttributeValue{&types.AttributeValueMemberS{Value: userID}}},
			":empty_list": &types.AttributeValueMemberL{Value: []types.AttributeValue{}},
		},
	})
	if err != nil {
		log.Printf("Error adding user to room: %v", err)
		http.Error(w, "Error joining room", http.StatusInternalServerError)
		return
	}

	result, err = h.DB.Client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(h.DB.RoomsTable),
		Key: map[string]types.AttributeValue{
			"roomId": &types.AttributeValueMemberS{Value: roomID},
		},
	})
	if err != nil {
		log.Printf("Error fetching updated room: %v", err)
		http.Error(w, "Error fetching updated room", http.StatusInternalServerError)
		return
	}

	err = attributevalue.UnmarshalMap(result.Item, &room)
	if err != nil {
		log.Printf("Error unmarshaling updated room: %v", err)
		http.Error(w, "Error processing room", http.StatusInternalServerError)
		return
	}

	log.Printf("User %s successfully joined room %s", userID, roomID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Successfully joined room",
		"room":    room,
	})
}
