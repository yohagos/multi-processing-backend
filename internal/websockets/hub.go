package websockets

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID        string
	UserID    string
	ChannelID string
	Conn      *websocket.Conn
	Send      chan []byte
}

type BroadcastMessage struct {
	ChannelID string
	UserID    string
	Type      string
	Payload   []byte
}

type Hub struct {
	clients    map[string]map[string]*Client
	broadcast  chan BroadcastMessage
	Register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]map[string]*Client),
		broadcast:  make(chan BroadcastMessage),
		Register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			if _, exists := h.clients[client.ChannelID]; !exists {
				h.clients[client.ChannelID] = make(map[string]*Client)
			}
			h.clients[client.ChannelID][client.ID] = client
			h.mu.Unlock()
			go client.writePump(h)
			go client.readPump(h)

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
	h.broadcast <- BroadcastMessage{
		ChannelID: channelID,
		Payload:   message,
	}
}

func (c *Client) writePump(h *Hub) {
	defer func() {
		c.Conn.Close()
		h.unregister <- c
	}()

	for message := range c.Send {
		w, err := c.Conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}

		if _, err := w.Write(message); err != nil {
			return
		}

		if err := w.Close(); err != nil {
			return
		}
	}

}

func (c *Client) readPump(h *Hub) {
	defer func() {
		h.unregister <- c
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		h.BroadcastToChannel(c.ChannelID, message)
	}
}

/* func (h *Hub) GetChannelClients(channelID string) []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var clients []*Client
	if channelClients, exists := h.clients[channelID]; exists {
		for _, client := range channelClients {
			clients = append(clients, client)
		}
	}
	return clients
} */
