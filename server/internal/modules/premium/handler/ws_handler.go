package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sovereign-fund/sovereign/internal/modules/premium/service"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // TODO: restrict in production
	},
}

type WSHandler struct {
	hub    *service.Hub
	logger *slog.Logger
}

func NewWSHandler(hub *service.Hub, logger *slog.Logger) *WSHandler {
	return &WSHandler{hub: hub, logger: logger}
}

func (h *WSHandler) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("ws upgrade failed", slog.String("error", err.Error()))
		return
	}

	userID, _ := c.Get("user_id")
	uid := ""
	if userID != nil {
		uid = userID.(string)
	}

	client := &service.Client{
		ID:     uuid.New().String(),
		UserID: uid,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Pairs:  make(map[string]bool),
	}

	h.hub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}
