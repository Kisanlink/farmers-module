package config

import (
	"github.com/Kisanlink/farmers-module/utils"
	"github.com/joho/godotenv"
	"os"
)

func LoadEnv() {

	err := godotenv.Load()
	if err != nil {
		utils.Log.Warn("⚠️ No .env file found, using system environment variables")
	}
}

func GetEnv(key string) string {
	return os.Getenv(key)
}
