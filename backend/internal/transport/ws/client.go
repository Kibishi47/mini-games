package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"nhooyr.io/websocket"

	"minigames-backend/internal/domain"
)

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	roomCode string
	userID   uuid.UUID
	sendChan chan []byte
	isClosed bool
	mu       sync.Mutex
}

func NewClient(hub *Hub, conn *websocket.Conn, roomCode string, userID uuid.UUID) *Client {
	return &Client{
		hub:      hub,
		conn:     conn,
		roomCode: roomCode,
		userID:   userID,
		sendChan: make(chan []byte, 64),
	}
}

func (c *Client) ReadPump(ctx context.Context) {
	defer func() {
		c.hub.Unregister(c)
		c.Close()
	}()

	for {
		typ, message, err := c.conn.Read(ctx)
		if err != nil {
			// Déconnexion ou fermeture
			break
		}
		if typ != websocket.MessageText {
			continue
		}

		var wsMsg domain.WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			c.SendError("Format de message invalide")
			continue
		}

		// Dispatcher
		c.hub.Dispatch(c, wsMsg)
	}
}

func (c *Client) WritePump(ctx context.Context) {
	ticker := time.NewTicker(20 * time.Second)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-c.sendChan:
			if !ok {
				_ = c.conn.Close(websocket.StatusNormalClosure, "")
				return
			}
			writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err := c.conn.Write(writeCtx, websocket.MessageText, msg)
			cancel()
			if err != nil {
				return
			}
		case <-ticker.C:
			// Ping / Heartbeat
			pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			err := c.conn.Ping(pingCtx)
			cancel()
			if err != nil {
				return
			}
		}
	}
}

func (c *Client) Send(msg domain.WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.isClosed {
		return
	}
	select {
	case c.sendChan <- data:
	default:
		// Buffer plein
	}
}

func (c *Client) SendError(msg string) {
	payload, _ := json.Marshal(map[string]string{"error": msg})
	c.Send(domain.WSMessage{
		Type:    "error",
		Payload: payload,
	})
}

func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.isClosed {
		c.isClosed = true
		close(c.sendChan)
	}
}
