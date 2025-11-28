package services

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/anant/realtime-pair-programming/internal/models"
	"github.com/gorilla/websocket"
)

type Client struct {
	ConnID   string
	UserID   string
	Username string
	RoomID   string
	Conn     *websocket.Conn
	Send     chan []byte
}

type RoomManager struct {
	rooms         map[string]map[string]*Client 
	broadcast     chan BroadcastMessage
	register      chan *Client
	unregister    chan *Client
	pendingLeaves map[string]*time.Timer 
	mu            sync.RWMutex
}

type BroadcastMessage struct {
	RoomID  string
	Message []byte
	Exclude string 
}

func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms:         make(map[string]map[string]*Client),
		broadcast:     make(chan BroadcastMessage, 256),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		pendingLeaves: make(map[string]*time.Timer),
	}
}

func (rm *RoomManager) BroadcastUserList(roomID string) {
	rm.mu.RLock()
	clients := rm.rooms[roomID]
	rm.mu.RUnlock()

	var userList []models.UserPresence
	seenUsers := make(map[string]bool)

	for _, client := range clients {
		if !seenUsers[client.UserID] {
			userList = append(userList, models.UserPresence{
				UserID:   client.UserID,
				Username: client.Username,
				Status:   "online",
			})
			seenUsers[client.UserID] = true
		}
	}

	msg := models.WSMessage{
		Type:    "user_list",
		Payload: userList,
	}
	data, _ := json.Marshal(msg)

	rm.BroadcastToRoom(roomID, data, "")
}

func (rm *RoomManager) Run() {
	for {
		select {
		case client := <-rm.register:
			rm.mu.Lock()
			if rm.rooms[client.RoomID] == nil {
				rm.rooms[client.RoomID] = make(map[string]*Client)
			}
			isFirstConnection := true
			for _, c := range rm.rooms[client.RoomID] {
				if c.UserID == client.UserID {
					isFirstConnection = false
					break
				}
			}

			rm.rooms[client.RoomID][client.ConnID] = client

			if timer, exists := rm.pendingLeaves[client.UserID]; exists {
				timer.Stop()
				delete(rm.pendingLeaves, client.UserID)
				isFirstConnection = false
			}

			rm.mu.Unlock()

			if isFirstConnection {
				joinMsg := models.WSMessage{
					Type: "user_joined",
					Payload: map[string]interface{}{
						"userId":   client.UserID,
						"username": client.Username,
					},
				}
				joinData, _ := json.Marshal(joinMsg)
				rm.BroadcastToRoom(client.RoomID, joinData, "")
			}

			go rm.BroadcastUserList(client.RoomID)

		case client := <-rm.unregister:
			rm.mu.Lock()
			if clients, ok := rm.rooms[client.RoomID]; ok {
				if _, ok := clients[client.ConnID]; ok {
					delete(clients, client.ConnID)
					close(client.Send)
					if len(clients) == 0 {
						delete(rm.rooms, client.RoomID)
					}
				}
			}

			isLastConnection := true
			if clients, ok := rm.rooms[client.RoomID]; ok {
				for _, c := range clients {
					if c.UserID == client.UserID {
						isLastConnection = false
						break
					}
				}
			}

			if isLastConnection {
				timer := time.AfterFunc(2*time.Second, func() {
					rm.mu.Lock()
					if _, exists := rm.pendingLeaves[client.UserID]; exists {
						delete(rm.pendingLeaves, client.UserID)

						leftMsg := models.WSMessage{
							Type: "user_left",
							Payload: map[string]interface{}{
								"userId":   client.UserID,
								"username": client.Username,
							},
						}
						leftData, _ := json.Marshal(leftMsg)
						rm.BroadcastToRoom(client.RoomID, leftData, "")
						go rm.BroadcastUserList(client.RoomID)
					}
					rm.mu.Unlock()
				})
				rm.pendingLeaves[client.UserID] = timer
			} else {
				go rm.BroadcastUserList(client.RoomID)
			}
			rm.mu.Unlock()

		case msg := <-rm.broadcast:
			rm.mu.RLock()
			if clients, ok := rm.rooms[msg.RoomID]; ok {
				for connID, client := range clients {
					if client.UserID != msg.Exclude {
						select {
						case client.Send <- msg.Message:
						default:
							close(client.Send)
							delete(clients, connID)
						}
					}
				}
			}
			rm.mu.RUnlock()
		}
	}
}

func (rm *RoomManager) RegisterClient(client *Client) {
	rm.register <- client
}

func (rm *RoomManager) UnregisterClient(client *Client) {
	rm.unregister <- client
}

func (rm *RoomManager) BroadcastToRoom(roomID string, message []byte, excludeUserID string) {
	rm.broadcast <- BroadcastMessage{
		RoomID:  roomID,
		Message: message,
		Exclude: excludeUserID,
	}
}

func (rm *RoomManager) GetRoomClients(roomID string) []*Client {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	clients := []*Client{}
	if roomClients, ok := rm.rooms[roomID]; ok {
		for _, client := range roomClients {
			clients = append(clients, client)
		}
	}
	return clients
}
