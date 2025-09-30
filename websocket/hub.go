package websocket

import (
	"echo-app/database"
	"log"
)

type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
}

var HubInstance = Hub{
	clients:    make(map[*Client]bool),
	register:   make(chan *Client),
	unregister: make(chan *Client),
	broadcast:  make(chan []byte),
}

func RunHub() {
	// Subscribe Redis → dorong ke broadcast channel
	go func() {
		log.Println("📡 Subscribing Redis channel chat...")
		sub := database.RedisClient.Subscribe(database.Ctx, "chat")
		ch := sub.Channel()
		for msg := range ch {
			log.Println("📥 Redis received:", msg.Payload)
			HubInstance.broadcast <- []byte(msg.Payload)
		}
	}()

	// Loop hub internal
	for {
		select {
		case client := <-HubInstance.register:
			HubInstance.clients[client] = true
			log.Printf("✅ Client %d connected", client.ID)

		case client := <-HubInstance.unregister:
			if _, ok := HubInstance.clients[client]; ok {
				delete(HubInstance.clients, client)
				close(client.Send)
				log.Printf("❌ Client %d disconnected", client.ID)
			}

		case message := <-HubInstance.broadcast:
			for client := range HubInstance.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(HubInstance.clients, client)
				}
			}
		}
	}
}
