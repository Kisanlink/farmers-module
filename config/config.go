package config

import (
	"os"

	"github.com/Kisanlink/farmers-module/utils"
	"github.com/joho/godotenv"
)

// filepath: c:\Users\Kaustubh\farmers-module\config\config.go
func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		utils.Log.Warn("⚠️ No .env file found, using system environment variables")
	} else {
		utils.Log.Info("✅ .env file loaded successfully")
	}
}

func GetEnv(key string) string {
	return os.Getenv(key)
}
