package ws

import "github.com/google/uuid"

type Hub struct {
	register   chan *Client
	unregister chan *Client
	
	toUser chan userMsg

	clients       map[*Client]struct{}
	clientsByUser map[uuid.UUID]map[*Client]struct{}
}

type userMsg struct {
	userID uuid.UUID
	msg    []byte
}

func NewHub() *Hub {
	return &Hub{
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		toUser:        make(chan userMsg, 256),
		clients:       make(map[*Client]struct{}),
		clientsByUser: make(map[uuid.UUID]map[*Client]struct{}),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.clients[c] = struct{}{}
			if h.clientsByUser[c.userID] == nil {
				h.clientsByUser[c.userID] = make(map[*Client]struct{})
			}
			h.clientsByUser[c.userID][c] = struct{}{}

		case c := <-h.unregister:
			h.removeClient(c)

		case m := <-h.toUser:
			h.publishToUser(m.userID, m.msg)
		}
	}
}

func (h *Hub) removeClient(c *Client) {
	if _, ok := h.clients[c]; !ok {
		return
	}
	delete(h.clients, c)

	if set := h.clientsByUser[c.userID]; set != nil {
		delete(set, c)
		if len(set) == 0 {
			delete(h.clientsByUser, c.userID)
		}
	}

	close(c.send)
}

func (h *Hub) publishToUser(userID uuid.UUID, msg []byte) {
	set := h.clientsByUser[userID]
	for c := range set {
		select {
		case c.send <- msg:
		default:
			h.removeClient(c)
		}
	}
}

func (h *Hub) PublishToUser(userID uuid.UUID, msg []byte) {
	h.toUser <- userMsg{userID: userID, msg: msg}
}
