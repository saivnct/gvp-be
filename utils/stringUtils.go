package utils

import (
	"encoding/hex"
	"errors"
	"fmt"
	"gbb.go/gvp/static"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"net/mail"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const numberBytes = "0123456789"

func CacheKey(collectionName, fieldName, fieldValue string) string {
	return collectionName + "_" + fieldName + "_" + fieldValue
}

func GenerateAuthenCode() string {
	return GenerateRandomNumberString(6)
}

func GenerateRandomNumberString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = numberBytes[rand.Intn(len(numberBytes))]
	}
	return string(b)
}

func GenerateRandomLetterString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)

	if err != nil {
		return nil, err
	}

	return b, nil
}

func GenerateRandomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func GenerateUUID() string {
	return uuid.New().String()
}

func GetPhoneFull(countryCode, phone string) string {
	return "+" + countryCode + phone
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func ParseJWTToken(jwtToken string) (*jwt.RegisteredClaims, error) {

	if len(jwtToken) == 0 {
		return nil, errors.New("Invalid token")
	}

	token, err := jwt.ParseWithClaims(jwtToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return static.JWTKey(), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("Cannot parse claims")
}

func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
