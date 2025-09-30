package websocket

import "github.com/gorilla/websocket"

type Client struct {
	ID    uint
	Rooms []uint
	Conn  *websocket.Conn
	Send  chan []byte
}
