package models

import (
	"sync"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type HexID struct {
	ID primitive.ObjectID `bson:"_id"`
}

type Connection struct {
	Socket *websocket.Conn
	mu     sync.Mutex
}

func (c *Connection) Send(message Message) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Socket.WriteJSON(message)
}
