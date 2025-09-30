package websocket

import (
	"echo-app/database"
	"echo-app/models"
	"echo-app/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func validateAndParseClaims(tokenString string) (*models.CustomClaims, error) {
	secret := utils.GetJwtSecret()
	token, err := jwt.ParseWithClaims(tokenString, &models.CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*models.CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token claims")
}

func ChatHandler(c echo.Context) error {
	log.Println("🌐 Incoming WebSocket connection...")

	// 🔑 Ambil token dari header atau query
	authHeader := c.Request().Header.Get("Authorization")
	var token string
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		token = strings.TrimSpace(authHeader[len("Bearer "):])
	} else {
		token = c.QueryParam("token")
		if token == "" {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "missing or invalid token"})
		}
	}

	// ✅ Validasi token
	claims, err := validateAndParseClaims(token)
	if err != nil {
		log.Println("❌ invalid token:", err)
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
	}
	userID := claims.ID
	roomIDs := claims.RoomsId

	// 🔄 Buat room default kalau kosong
	if len(roomIDs) == 0 {
		var newRoom models.Rooms
		if err := database.DB.FirstOrCreate(&newRoom, models.Rooms{Name: fmt.Sprintf("Room-%d", userID)}).Error; err != nil {
			log.Printf("❌ gagal buat room: %v", err)
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to create room"})
		}
		roomIDs = append(roomIDs, newRoom.ID)
	}

	// 🔄 Upgrade ke WebSocket
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Println("❌ upgrade error:", err)
		return err
	}

	client := &Client{
		ID:    userID,
		Rooms: roomIDs,
		Conn:  ws,
		Send:  make(chan []byte, 256),
	}

	// Register client ke Hub
	HubInstance.register <- client

	// Kirim history hanya ke client baru
	for _, roomID := range roomIDs {
		var history []struct {
			ID         uint      `json:"id"`
			RoomID     uint      `json:"room_id"`
			SenderID   uint      `json:"sender_id"`
			Content    string    `json:"content"`
			MsgType    string    `json:"type"`
			CreatedAt  time.Time `json:"created_at"`
			SenderName string    `json:"sender_name"`
		}
		err := database.DB.Table("messages AS m").
			Select(`m.id, m.room_id, m.sender_id, m.content, m."type", m.created_at, p.full_name AS sender_name`).
			Joins("LEFT JOIN profiles p ON p.user_id = m.sender_id").
			Where("m.room_id = ?", roomID).
			Order("m.created_at ASC").
			Scan(&history).Error
		if err != nil {
			log.Printf("❌ DB error ambil history room %d: %v", roomID, err)
			continue
		}

		payload, _ := json.Marshal(echo.Map{
			"type":     "history",
			"room_id":  roomID,
			"messages": history,
		})
		client.Send <- payload
	}

	// Jalankan pump
	go WritePump(client)
	ReadPump(client)

	return nil
}
