package database

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var (
	RedisClient *redis.Client
	Ctx         = context.Background()
)

func ConnectRedis() {
	addr := os.Getenv("REDIS_ADDR") // contoh: redis:6379
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // isi jika ada password
		DB:       0,
	})

	_, err := RedisClient.Ping(Ctx).Result()
	if err != nil {
		log.Fatal("❌ Redis connection failed:", err)
	}
	log.Println("✅ Redis connected at", addr)
}
