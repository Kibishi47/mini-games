package ws

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	conn   *websocket.Conn
	send   chan []byte
	userID uuid.UUID
}

func NewClient(conn *websocket.Conn, userID uuid.UUID) *Client {
	return &Client{
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: userID,
	}
}
