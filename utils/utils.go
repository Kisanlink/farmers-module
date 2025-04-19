package utils

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"
)

// Generate10DigitId generates a random 10-digit number as a string
func Generate10DigitId() string {
	rand.Seed(time.Now().UnixNano())
	min := 1000000000 // Smallest 10-digit number
	max := 9999999999 // Largest 10-digit number
	id := rand.Intn(max-min+1) + min
	return fmt.Sprintf("%010d", id) // Ensure it's always 10 digits
}

// LogError logs an error message
func LogError(message string) {
	log.Println(message)
}

func GenerateCycleId() string {
	uniqueNumber := rand.Intn(90000) + 10000 // always 5 digits
	return "CYCLE" + strconv.Itoa(uniqueNumber)
}
