package auth

import (
	"time"
	"fmt"
	"strings"
	"net/http"
	"crypto/rand"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func MakeRefreshToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)
	return token, nil
}

func GetBearerToken(header http.Header) (string, error) {
	bearerToken := header.Get("Authorization")
	if bearerToken == "" {
		return "", fmt.Errorf("Authorization header not found")
	}

	token, ok := strings.CutPrefix(bearerToken, "Bearer ")
	if !ok {
		return "", fmt.Errorf("Unable to find 'Bearer ' prefix")
	}
	return token, nil
}

func GetAPIKey(header http.Header) (string, error) {
	apiKey := header.Get("Authorization")
	if apiKey == "" {
		return "", fmt.Errorf("Authorization header not found")
	}

	key, ok := strings.CutPrefix(apiKey, "ApiKey ")
	if !ok {
		return "", fmt.Errorf("Unable to find 'ApiKey ', prefix")
	}
	return key, nil


}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func CheckPasswordHash(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {

	now := time.Now().UTC()
	claim := jwt.RegisteredClaims{
		Issuer: "chirpy",
		IssuedAt: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
		Subject: userID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	s, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return s, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {

	claims := jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(tokenString, &claims,
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(tokenSecret), nil
		},
	)
	if err != nil {
		return uuid.UUID{}, err
	}

	if token.Valid {
		strUID := token.Claims.(*jwt.RegisteredClaims).Subject
		UUID := uuid.MustParse(strUID)
		return UUID, nil
	} else {
		return uuid.UUID{}, fmt.Errorf("invalid token")
	}

}
