package ws

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"nhooyr.io/websocket"

	"minigames-backend/internal/domain"
)

type Client struct {
	hub              *Hub
	conn             *websocket.Conn
	roomCode         string
	userID           uuid.UUID
	sessionToken     string
	sendChan         chan []byte
	isClosed         bool
	isVoluntaryLeave bool
	mu               sync.Mutex
}

func NewClient(hub *Hub, conn *websocket.Conn, roomCode string, userID uuid.UUID, sessionToken string) *Client {
	return &Client{
		hub:          hub,
		conn:         conn,
		roomCode:     roomCode,
		userID:       userID,
		sessionToken: sessionToken,
		sendChan:     make(chan []byte, 64),
	}
}

func (c *Client) RoomCode() string {
	return c.roomCode
}

func (c *Client) UserID() uuid.UUID {
	return c.userID
}

func (c *Client) SessionToken() string {
	return c.sessionToken
}

func (c *Client) SetVoluntaryLeave(val bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.isVoluntaryLeave = val
}

func (c *Client) IsVoluntaryLeave() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.isVoluntaryLeave
}

func (c *Client) ReadPump(ctx context.Context) {
	defer func() {
		c.hub.Unregister(c)
		c.Close()
	}()

	for {
		typ, message, err := c.conn.Read(ctx)
		if err != nil {
			break
		}
		if typ != websocket.MessageText {
			continue
		}

		var wsMsg domain.WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			continue
		}

		c.hub.HandleMessage(c, wsMsg)
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
				return
			}
			writeCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			err := c.conn.Write(writeCtx, websocket.MessageText, msg)
			cancel()
			if err != nil {
				return
			}
		case <-ticker.C:
			// Ping régulier pour maintenir la connexion active
			writeCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			err := c.conn.Ping(writeCtx)
			cancel()
			if err != nil {
				return
			}
			// Rafraîchir le heartbeat d'activité
			c.hub.TouchPlayer(c)
		}
	}
}

func (c *Client) Send(msg domain.WSMessage) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isClosed {
		return
	}

	bytes, err := json.Marshal(msg)
	if err != nil {
		return
	}

	select {
	case c.sendChan <- bytes:
	default:
		// Tampon saturé
	}
}

func (c *Client) SendError(message string) {
	payload, _ := json.Marshal(map[string]string{"error": message})
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
		_ = c.conn.Close(websocket.StatusNormalClosure, "Fermeture normale")
	}
}
