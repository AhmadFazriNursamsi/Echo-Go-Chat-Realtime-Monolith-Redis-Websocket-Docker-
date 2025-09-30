package main

import (
	"echo-app/database"
	"echo-app/handlers"
	middlewareLokal "echo-app/middleware" // <- ini package lokal Anda
	"echo-app/models"

	"log"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"

	lokal "echo-app/websocket"

	// gorilla "github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Validator custom
type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

// var upgrader = gorilla.Upgrader{
// 	ReadBufferSize:  1024,
// 	WriteBufferSize: 1024,
// 	CheckOrigin: func(r *http.Request) bool {
// 		// TODO: batasi origin di production
// 		return true
// 	},
// }

func main() {
	// Load env manual (atau pakai godotenv)
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ Tidak bisa load .env file, pakai env system")
	}

	// Init DB
	database.Connect()
	database.ConnectRedis()

	database.DB.AutoMigrate(&models.User{}, &models.Role{}, &models.Profile{}, &models.Messages{}, &models.Rooms{})

	// Init Echo
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}

	// Middleware bawaan
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	go lokal.RunHub()

	// Routes
	auth := e.Group("/api/auth")
	auth.POST("/register", handlers.Register)
	auth.POST("/login", handlers.Login)
	auth.PUT("/change-password", handlers.ChangePassword, middlewareLokal.JWTMiddleware())
	auth.PUT("/change-email", handlers.ChangeEmail, middlewareLokal.JWTMiddleware())
	auth.POST("/forgot-password", handlers.ForgotPassword)
	auth.POST("/reset-password", handlers.ResetPassword)

	//websocket
	e.GET("/ws", lokal.ChatHandler)

	api := e.Group("/api/v1")
	api.POST("/roles", handlers.CreateRole)
	api.GET("/roles", handlers.GetRoles)
	api.GET("/roles/:id", handlers.GetRolesByid)

	api.POST("/profiles", handlers.CreateProfile)
	api.GET("/profiles", handlers.GetProfiles)
	api.GET("/profiles/:id", handlers.GetProfileByID)
	api.DELETE("/profiles/:id", handlers.DeleteProfiles)
	api.PUT("/profiles/:id", handlers.UpdateProfile)
	api.GET("/me", handlers.GetMyProfile, middlewareLokal.JWTMiddleware())

	// Protected routes
	api.GET("/profile", func(c echo.Context) error {
		user := c.Get("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		return c.JSON(200, echo.Map{
			"user_id": claims["user_id"],
		})
	}, middlewareLokal.JWTMiddleware())

	// Start server
	log.Println("🚀 Server running at :8080")
	e.Logger.Fatal(e.Start(":8080"))

}
