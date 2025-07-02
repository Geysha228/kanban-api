package util

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"kanban-api/config"
	"math/big"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var cfgEmail config.EmailConfirm
var auth smtp.Auth

const secretTokenKey = `AxiTech`

func CreateEmailCode() (*big.Int, error) {
	max := big.NewInt(900000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return nil, err
	}
	n = n.Add(n, big.NewInt(100000))
	return n, nil
}

func SetConfigEmail(){
	cfgEmail = config.GetConfigEmail()
	auth = smtp.PlainAuth("", cfgEmail.EmailFrom, cfgEmail.PasswordFrom, cfgEmail.SmtpHost)
}

func CreateEmailMessage(code big.Int, email string, description string) string {
	msg := fmt.Sprintf("Subject: %s\r\n", "Confirmation your email in Kanban-desk") +
    fmt.Sprintf("From: %s\r\n", cfgEmail.EmailFrom) +
    fmt.Sprintf("To: %s\r\n", email) +
    "MIME-Version: 1.0\r\n" +
    "Content-Type: text/plain; charset=\"UTF-8\"\r\n" +
    "\r\n" + // пустая строка отделяет заголовки от тела письма
    fmt.Sprintf("Your confirmation code: %v\n", code.String()) +
	description
	return msg
}

func SendMail(message string, emailTo string) error {
	err := smtp.SendMail(cfgEmail.SmtpHost + ":" + cfgEmail.SmtpPort, auth, cfgEmail.EmailFrom, []string{emailTo}, []byte(message))
	return err
}

func DecodeJSONBody[T any](r *http.Request)(T, error){
	var body T
	decoder := json.NewDecoder(r.Body)
    decoder.DisallowUnknownFields()
    err := decoder.Decode(&body)
    return body, err
}

func CreateJWT(hours int, userID int)(string, error){
	secretKey := []byte(secretTokenKey)
	claims := jwt.MapClaims{
		"user_id": userID,
        "exp":     time.Now().Add(time.Duration(hours) * time.Hour).Unix(),
        "iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(secretKey)
}

func ParseJWT(tokenString string) (int, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
    	// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretTokenKey), nil
	})

	
	if err != nil {
		return 0, err
    }

    // Проверяем валидность токена
    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        // Извлекаем user_id (предполагается, что он int)
        if userIDFloat, ok := claims["user_id"].(float64); ok {
            userID := int(userIDFloat)
            return userID, nil
        }
        return 0, fmt.Errorf("user_id not found in token claims")
    } else {
        return 0, fmt.Errorf("invalid token")
    }
}

func GetTokenFromRequest(r *http.Request) (string, error) {
    authHeader := r.Header.Get("Authorization")
    if authHeader == "" {
        return "", errors.New("authorization header is missing")
    }

    parts := strings.SplitN(authHeader, " ", 2)
    if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
        return "", errors.New("authorization header format must be Bearer {token}")
    }

    tokenString := parts[1]
    return tokenString, nil
}
