package websocket

import (
	"echo-app/database"
	"echo-app/models"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

// ReadPump: terima pesan dari client → simpan DB (untuk type "message")
// atau update read receipt (untuk type "read") → publish Redis → Hub broadcast
func ReadPump(c *Client) {
	// Setup read deadline & pong handler
	c.Conn.SetReadLimit(1024 * 64)
	_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	defer func() {
		HubInstance.unregister <- c
		_ = c.Conn.Close()
	}()

	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println("❌ Read error:", err)
			break
		}

		// Parse JSON masuk
		var incoming struct {
			RoomID  uint   `json:"room_id"`
			Content string `json:"content"`
			Type    string `json:"type"`    // "message" | "read"
			UserID  uint   `json:"user_id"` // untuk "read"
		}
		if err := json.Unmarshal(msg, &incoming); err != nil {
			log.Println("❌ Invalid message format:", err)
			continue
		}

		// Validasi RoomID minimal
		if incoming.RoomID == 0 {
			log.Println("❌ Missing room_id")
			continue
		}

		// --- READ RECEIPT ---
		if incoming.Type == "read" {
			// Hindari duplikat: remove dulu, lalu append
			err := database.DB.Model(&models.Messages{}).
				Where("room_id = ?", incoming.RoomID).
				Updates(map[string]interface{}{
					"read_by": gorm.Expr("array_append(array_remove(COALESCE(read_by, '{}'), ?), ?)", incoming.UserID, incoming.UserID),
				}).Error
			if err != nil {
				log.Println("❌ gagal update read receipt:", err)
				continue
			}

			receipt := struct {
				Type   string `json:"type"`
				RoomID uint   `json:"room_id"`
				UserID uint   `json:"user_id"`
				Status string `json:"status"`
			}{
				Type:   "receipt",
				RoomID: incoming.RoomID,
				UserID: incoming.UserID,
				Status: "read",
			}
			payload, _ := json.Marshal(receipt)

			// Broadcast lokal & publish Redis
			HubInstance.broadcast <- payload
			err = database.RedisClient.Publish(database.Ctx, "chat", string(payload)).Err()
			if err != nil {
				log.Println("❌ Redis publish error:", err)
			} else {
				log.Println("✅ Redis publish OK:", string(payload))
			}
			continue
		}

		// --- PESAN BARU ---
		if incoming.Type == "" {
			incoming.Type = "message"
		}
		message := models.Messages{
			RoomId:   incoming.RoomID, // per struct kamu
			SenderId: c.ID,
			Content:  incoming.Content,
			MsgType:  "message",
			// CreatedAt akan otomatis kalau pakai gorm.Model; kalau custom tambahkan time.Now()
		}
		if err := database.DB.Create(&message).Error; err != nil {
			log.Println("❌ DB save error:", err)
			continue
		}

		// Tandai delivered_to (unique)
		_ = database.DB.Model(&message).
			Update("delivered_to", gorm.Expr("array_append(array_remove(COALESCE(delivered_to, '{}'), ?), ?)", c.ID, c.ID)).Error

		// Angkat payload lengkap (join profile untuk sender_name)
		var out struct {
			ID         uint      `json:"id"`
			RoomID     uint      `json:"room_id"`
			SenderID   uint      `json:"sender_id"`
			Content    string    `json:"content"`
			MsgType    string    `json:"type"`
			CreatedAt  time.Time `json:"created_at"`
			SenderName string    `json:"sender_name"`
		}

		// Penting: alias tabel benar dan kutip kolom "type"
		if err := database.DB.Table("messages AS m").
			Select(`m.id, m.room_id, m.sender_id, m.content, m."type", m.created_at, COALESCE(p.full_name, '') AS sender_name`).
			Joins("LEFT JOIN profiles p ON p.user_id = m.sender_id").
			Where("m.id = ?", message.ID).
			Scan(&out).Error; err != nil {
			// fallback minimal tanpa join
			out = struct {
				ID         uint      `json:"id"`
				RoomID     uint      `json:"room_id"`
				SenderID   uint      `json:"sender_id"`
				Content    string    `json:"content"`
				MsgType    string    `json:"type"`
				CreatedAt  time.Time `json:"created_at"`
				SenderName string    `json:"sender_name"`
			}{
				ID:         message.ID,
				RoomID:     message.RoomId,
				SenderID:   message.SenderId,
				Content:    message.Content,
				MsgType:    "message",
				CreatedAt:  time.Now(),
				SenderName: "",
			}
		}

		payload, _ := json.Marshal(out)

		// Broadcast lokal
		// HubInstance.broadcast <- payload

		// Publish Redis (SATU KALI saja, jangan double)
		if err := database.RedisClient.Publish(database.Ctx, "chat", string(payload)).Err(); err != nil {
			log.Println("❌ Redis publish error:", err)
		} else {
			log.Println("✅ Redis publish OK:", string(payload))
		}
	}
}

// WritePump: kirim pesan ke client + heartbeat ping
func WritePump(client *Client) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			_ = client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = client.Conn.WriteMessage(websocket.CloseMessage, nil)
				log.Println("Channel closed for client")
				return
			}

			log.Println("Send to client:", string(message))
			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)
			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			_ = client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
