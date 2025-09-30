package auth

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/patrickmn/go-cache"
)

// TokenValidator handles JWT token validation with caching
type TokenValidator struct {
	publicKey *rsa.PublicKey
	secret    []byte
	cache     *cache.Cache
	issuer    string
	audience  string
}

// TokenClaims represents the JWT token claims
type TokenClaims struct {
	UserID      string   `json:"sub"`
	OrgID       string   `json:"org_id"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	TokenType   string   `json:"token_type"`
	jwt.RegisteredClaims
}

// NewTokenValidator creates a new token validator
func NewTokenValidator(secret string, publicKeyPEM []byte) (*TokenValidator, error) {
	var publicKey *rsa.PublicKey
	if len(publicKeyPEM) > 0 {
		key, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key: %w", err)
		}
		publicKey = key
	}

	return &TokenValidator{
		publicKey: publicKey,
		secret:    []byte(secret),
		cache:     cache.New(5*time.Minute, 10*time.Minute),
		issuer:    "aaa-service",
		audience:  "farmers-module",
	}, nil
}

// ValidateToken validates a JWT token and returns claims
func (tv *TokenValidator) ValidateToken(ctx context.Context, tokenString string) (*TokenClaims, error) {
	if tokenString == "" {
		return nil, errors.New("token is required")
	}

	// Check cache first
	if cached, found := tv.cache.Get(tokenString); found {
		if claims, ok := cached.(*TokenClaims); ok {
			// Check if still valid
			if claims.ExpiresAt != nil && claims.ExpiresAt.After(time.Now()) {
				return claims, nil
			}
			tv.cache.Delete(tokenString)
		}
	}

	// Parse and validate token
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		switch token.Method.(type) {
		case *jwt.SigningMethodRSA:
			if tv.publicKey == nil {
				return nil, errors.New("RSA public key not configured")
			}
			return tv.publicKey, nil
		case *jwt.SigningMethodHMAC:
			return tv.secret, nil
		default:
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
	})

	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	// Validate issuer if set
	if claims.Issuer != "" && claims.Issuer != tv.issuer {
		return nil, fmt.Errorf("invalid issuer: expected %s, got %s", tv.issuer, claims.Issuer)
	}

	// Validate audience if set
	if len(claims.Audience) > 0 && !contains(claims.Audience, tv.audience) {
		return nil, fmt.Errorf("invalid audience: token not intended for %s", tv.audience)
	}

	// Cache valid token
	if claims.ExpiresAt != nil {
		tv.cache.Set(tokenString, claims, time.Until(claims.ExpiresAt.Time))
	}

	return claims, nil
}

// ExtractTokenFromHeader extracts Bearer token from Authorization header
func (tv *TokenValidator) ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is empty")
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || !strings.HasPrefix(authHeader, bearerPrefix) {
		return "", errors.New("invalid authorization header format")
	}

	return authHeader[len(bearerPrefix):], nil
}

// ValidateTokenFromHeader extracts and validates token from Authorization header
func (tv *TokenValidator) ValidateTokenFromHeader(ctx context.Context, authHeader string) (*TokenClaims, error) {
	token, err := tv.ExtractTokenFromHeader(authHeader)
	if err != nil {
		return nil, err
	}

	return tv.ValidateToken(ctx, token)
}

// contains checks if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
