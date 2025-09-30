package utils

import (
	"encoding/base64"
	"fmt"
	"os"
)

// getJwtSecret membaca JWT_SECRET dari environment.
// Jika value base64 valid → decode, kalau tidak → pakai raw string.
func GetJwtSecret() []byte {
	raw := os.Getenv("JWT_SECRET")
	if raw == "" {
		fmt.Println("⚠️  JWT_SECRET tidak ditemukan, pakai default.")
		return []byte("default-secret")
	}

	// coba decode base64
	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err == nil {
		fmt.Println("🔑 JWT_SECRET terbaca sebagai base64 (decoded).")
		return decoded
	}

	// fallback → plain string
	fmt.Println("🔑 JWT_SECRET pakai raw string.")
	return []byte(raw)
}
