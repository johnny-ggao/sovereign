package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sovereign-fund/sovereign/internal/modules/premium/dto"
	"github.com/sovereign-fund/sovereign/internal/modules/premium/model"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	maxMsgSize = 512
)

type Hub struct {
	mu         sync.RWMutex
	clients    map[string]*Client
	register   chan *Client
	unregister chan *Client
	logger     *slog.Logger
}

type Client struct {
	ID     string
	UserID string
	Conn   *websocket.Conn
	Send   chan []byte
	Pairs  map[string]bool
	hub    *Hub
}

func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		logger:     logger,
	}
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			h.mu.Lock()
			for id, client := range h.clients {
				close(client.Send)
				delete(h.clients, id)
			}
			h.mu.Unlock()
			h.logger.Info("ws hub stopped")
			return

		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()
			h.logger.Debug("ws client connected",
				slog.String("client_id", client.ID),
				slog.Int("total", h.ClientCount()),
			)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.Send)
			}
			h.mu.Unlock()
			h.logger.Debug("ws client disconnected",
				slog.String("client_id", client.ID),
				slog.Int("total", h.ClientCount()),
			)
		}
	}
}

func (h *Hub) Register(client *Client) {
	client.hub = h
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (h *Hub) BroadcastTick(snapshot model.PremiumSnapshot) {
	msg := dto.WSMessage{
		Type: "tick",
		Data: dto.PremiumTickResponse{
			Pair:              snapshot.Pair,
			KoreanPrice:       snapshot.KoreanPrice,
			GlobalPrice:       snapshot.GlobalPrice,
			PremiumPct:        snapshot.PremiumPct,
			ReversePremiumPct: snapshot.ReversePremiumPct,
			SourceKR:          snapshot.SourceKR,
			SourceGL:          snapshot.SourceGL,
			Latencies:         snapshot.Latencies,
			Timestamp:         snapshot.Timestamp.Format(time.RFC3339),
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		h.logger.Error("failed to marshal tick", slog.String("error", err.Error()))
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		if len(client.Pairs) > 0 && !client.Pairs[snapshot.Pair] {
			continue
		}
		select {
		case client.Send <- data:
		default:
			go func(c *Client) {
				select {
				case h.unregister <- c:
				case <-time.After(5 * time.Second):
				}
			}(client)
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMsgSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			return
		}

		var sub dto.WSSubscribeMessage
		if err := json.Unmarshal(message, &sub); err != nil {
			continue
		}

		switch sub.Action {
		case "subscribe":
			for _, p := range sub.Pairs {
				c.Pairs[p] = true
			}
			resp, _ := json.Marshal(dto.WSMessage{
				Type: "subscribed",
				Data: map[string]interface{}{"pairs": sub.Pairs},
			})
			select {
			case c.Send <- resp:
			default:
			}

		case "unsubscribe":
			for _, p := range sub.Pairs {
				delete(c.Pairs, p)
			}
		}
	}
}
