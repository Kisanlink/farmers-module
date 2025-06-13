package utils

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
)

func Generate10DigitId() string {
	min := int64(1000000000)
	max := int64(9999999999)
	nBig, err := rand.Int(rand.Reader, big.NewInt(max-min+1))
	if err != nil {
		LogError(fmt.Sprintf("failed to generate random number: %v", err))
		return ""
	}
	id := nBig.Int64() + min
	return fmt.Sprintf("%010d", id)
}

// LogError logs an error message
func LogError(message string) {
	log.Println(message)
}

// GenerateCycleId generates a random 5-digit number prefixed with "CYCLE"
func GenerateCycleId() string {
	num, _ := rand.Int(rand.Reader, big.NewInt(90000))
	id := num.Int64() + 10000 // ensures 5 digits
	return fmt.Sprintf("CYCLE%d", id)
}

// GenerateActId generates a random 7-digit number prefixed with "ACT"
func GenerateActId() string {
	num, _ := rand.Int(rand.Reader, big.NewInt(9000000))
	id := num.Int64() + 1000000 // ensures 7 digits
	return fmt.Sprintf("ACT%d", id)
}
