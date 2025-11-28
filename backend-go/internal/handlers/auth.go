package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/anant/realtime-pair-programming/internal/auth"
	"github.com/anant/realtime-pair-programming/internal/db"
	"github.com/anant/realtime-pair-programming/internal/models"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"github.com/aws/aws-sdk-go-v2/aws"
)

type AuthHandler struct {
	DB *db.DynamoDB
}

func NewAuthHandler(database *db.DynamoDB) *AuthHandler {
	return &AuthHandler{DB: database}
}

func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var req models.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" || req.Username == "" {
		http.Error(w, "Email, username and password are required", http.StatusBadRequest)
		return
	}

	result, err := h.DB.Client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String(h.DB.UsersTable),
		IndexName:              aws.String("EmailIndex"),
		KeyConditionExpression: aws.String("email = :email"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":email": &types.AttributeValueMemberS{Value: req.Email},
		},
	})
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if result.Count > 0 {
		http.Error(w, "User with this email already exists", http.StatusConflict)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error processing password", http.StatusInternalServerError)
		return
	}

	user := models.User{
		UserID:         uuid.New().String(),
		Username:       req.Username,
		Email:          req.Email,
		HashedPassword: string(hashedPassword),
		CreatedAt:      time.Now(),
		LastSeen:       time.Now(),
	}

	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	_, err = h.DB.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(h.DB.UsersTable),
		Item:      item,
	})
	if err != nil {
		http.Error(w, "Error saving user", http.StatusInternalServerError)
		return
	}

	token, err := auth.GenerateToken(user.UserID, user.Username, user.Email)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	response := models.AuthResponse{
		Token:  token,
		UserID: user.UserID,
		User:   user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := h.DB.Client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String(h.DB.UsersTable),
		IndexName:              aws.String("EmailIndex"),
		KeyConditionExpression: aws.String("email = :email"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":email": &types.AttributeValueMemberS{Value: req.Email},
		},
	})
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if result.Count == 0 {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	var user models.User
	err = attributevalue.UnmarshalMap(result.Items[0], &user)
	if err != nil {
		http.Error(w, "Error processing user data", http.StatusInternalServerError)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(req.Password))
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	_, err = h.DB.Client.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName: aws.String(h.DB.UsersTable),
		Key: map[string]types.AttributeValue{
			"userId": &types.AttributeValueMemberS{Value: user.UserID},
		},
		UpdateExpression: aws.String("SET lastSeen = :now"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":now": &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
		},
	})

	token, err := auth.GenerateToken(user.UserID, user.Username, user.Email)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	response := models.AuthResponse{
		Token:  token,
		UserID: user.UserID,
		User:   user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
