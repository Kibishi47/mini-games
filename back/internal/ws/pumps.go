package ws

func (c *Client) writePump() {
	defer c.conn.Close()

	for msg := range c.send {
		if err := c.conn.WriteMessage(1, msg); err != nil {
			return
		}
	}
}
