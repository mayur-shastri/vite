package auth

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// IMPORTANT: In a real application, this secret key should be loaded from
// a secure source like an environment variable, not hardcoded.
var jwtSecretKey = []byte(os.Getenv("JWT_SECRET_KEY"))

// AppClaims represents the custom claims for our JWT.
// It includes the standard registered claims and our own UserID.
type AppClaims struct {
	UserID uint `json:"userID"`
	jwt.RegisteredClaims
}

// GenerateToken creates a new JWT for a given user ID.
func GenerateToken(userID uint) (string, error) {
	// Set the token's expiration time.
	// For this example, we'll set it to 24 hours.
	expirationTime := time.Now().Add(24 * time.Hour)

	// Create the claims
	claims := &AppClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			// Set the expiration time
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			// Set the issued at time
			IssuedAt: jwt.NewNumericDate(time.Now()),
			// Set the issuer
			Issuer: "balkanid-filevault",
		},
	}

	// Create a new token object, specifying the signing method and the claims.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with our secret key to get the complete, signed token string.
	tokenString, err := token.SignedString(jwtSecretKey)
	if err != nil {
		return "", fmt.Errorf("could not sign token: %w", err)
	}

	return tokenString, nil
}

func ValidateToken(tokenString string) (*AppClaims, error) {
	claims := &AppClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("could not parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
