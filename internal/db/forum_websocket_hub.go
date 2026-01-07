package db

import (
	"sync"

	"github.com/gorilla/websocket"
)

// WebSocket Hub

type Client struct {
	ID        string
	UserId    string
	ChannelID string
	Conn      *websocket.Conn
	Send      chan []byte
}

type Message struct {
	ChannelID string `json:"channel_id"`
	UserID    string `json:"user_id"`
	Type      string `json:"type"`
	Payload   []byte `json:"payload"`
}

type Hub struct {
	clients    map[string]map[string]*Client
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]map[string]*Client),
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if _, exists := h.clients[client.ChannelID]; !exists {
				h.clients[client.ChannelID] = make(map[string]*Client)
			}
			h.clients[client.ChannelID][client.ID] = client
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if channelClients, exists := h.clients[client.ChannelID]; exists {
				if _, clientExists := channelClients[client.ID]; clientExists {
					delete(channelClients, client.ID)
					close(client.Send)
					if len(channelClients) == 0 {
						delete(h.clients, client.ChannelID)
					}
				}
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.mu.RLock()
			if channelClients, exists := h.clients[message.ChannelID]; exists {
				for _, client := range channelClients {
					select {
					case client.Send <- message.Payload:
					default:
						close(client.Send)
						delete(channelClients, client.ID)
					}
				}
			}
			h.mu.RUnlock()
		}

	}
}

func (h *Hub) BroadcastToChannel(channelID string, message []byte) {
	h.broadcast <- Message{
		ChannelID: channelID,
		Payload: message,
	}
}

func (h *Hub) GetChannelClients(channelID string) []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var clients []*Client
	if channelClients, exists := h.clients[channelID]; exists {
		for _, client := range channelClients {
			clients = append(clients, client)
		}
	}
	return clients
}
