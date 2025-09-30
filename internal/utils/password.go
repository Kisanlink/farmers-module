package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const (
	// Character sets for password generation
	upperChars   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lowerChars   = "abcdefghijklmnopqrstuvwxyz"
	digitChars   = "0123456789"
	specialChars = "!@#$%^&*()_+-=[]{}|;:,.<>?"

	// Minimum requirements
	minLength         = 12
	minUpper          = 2
	minLower          = 2
	minDigits         = 2
	minSpecial        = 2
	defaultBcryptCost = 10
)

// PasswordGenerator handles secure password generation and hashing
type PasswordGenerator struct {
	length int
	cost   int
}

// NewPasswordGenerator creates a new password generator
func NewPasswordGenerator() *PasswordGenerator {
	return &PasswordGenerator{
		length: 16,
		cost:   defaultBcryptCost,
	}
}

// GenerateSecurePassword generates a cryptographically secure password
func (pg *PasswordGenerator) GenerateSecurePassword() (string, error) {
	// Ensure minimum requirements are met
	password := make([]byte, 0, pg.length)

	// Add required characters
	for i := 0; i < minUpper; i++ {
		char, err := randomChar(upperChars)
		if err != nil {
			return "", err
		}
		password = append(password, char)
	}

	for i := 0; i < minLower; i++ {
		char, err := randomChar(lowerChars)
		if err != nil {
			return "", err
		}
		password = append(password, char)
	}

	for i := 0; i < minDigits; i++ {
		char, err := randomChar(digitChars)
		if err != nil {
			return "", err
		}
		password = append(password, char)
	}

	for i := 0; i < minSpecial; i++ {
		char, err := randomChar(specialChars)
		if err != nil {
			return "", err
		}
		password = append(password, char)
	}

	// Fill remaining length with random characters
	allChars := upperChars + lowerChars + digitChars + specialChars
	for len(password) < pg.length {
		char, err := randomChar(allChars)
		if err != nil {
			return "", err
		}
		password = append(password, char)
	}

	// Shuffle the password
	shuffled, err := shuffle(password)
	if err != nil {
		return "", err
	}

	return string(shuffled), nil
}

// HashPassword hashes a password using bcrypt
func (pg *PasswordGenerator) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), pg.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword verifies a password against its hash
func (pg *PasswordGenerator) VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// GenerateAPIKey generates a secure API key
func (pg *PasswordGenerator) GenerateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Encode to URL-safe base64
	key := base64.URLEncoding.EncodeToString(bytes)
	key = strings.TrimRight(key, "=") // Remove padding

	return key, nil
}

// randomChar returns a random character from the given charset
func randomChar(charset string) (byte, error) {
	max := big.NewInt(int64(len(charset)))
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0, err
	}
	return charset[n.Int64()], nil
}

// shuffle randomly shuffles a byte slice using Fisher-Yates algorithm
func shuffle(data []byte) ([]byte, error) {
	result := make([]byte, len(data))
	copy(result, data)

	for i := len(result) - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return nil, err
		}
		result[i], result[j.Int64()] = result[j.Int64()], result[i]
	}

	return result, nil
}

// ValidatePasswordStrength validates if a password meets security requirements
func (pg *PasswordGenerator) ValidatePasswordStrength(password string) error {
	if len(password) < minLength {
		return fmt.Errorf("password must be at least %d characters long", minLength)
	}

	var upperCount, lowerCount, digitCount, specialCount int

	for _, char := range password {
		switch {
		case strings.ContainsRune(upperChars, char):
			upperCount++
		case strings.ContainsRune(lowerChars, char):
			lowerCount++
		case strings.ContainsRune(digitChars, char):
			digitCount++
		case strings.ContainsRune(specialChars, char):
			specialCount++
		}
	}

	if upperCount < minUpper {
		return fmt.Errorf("password must contain at least %d uppercase characters", minUpper)
	}
	if lowerCount < minLower {
		return fmt.Errorf("password must contain at least %d lowercase characters", minLower)
	}
	if digitCount < minDigits {
		return fmt.Errorf("password must contain at least %d digits", minDigits)
	}
	if specialCount < minSpecial {
		return fmt.Errorf("password must contain at least %d special characters", minSpecial)
	}

	return nil
}
