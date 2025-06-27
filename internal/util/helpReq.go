package util

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"kanban-api/config"
	"math/big"
	"net/http"
	"net/smtp"
)

var cfgEmail config.EmailConfirm
var auth smtp.Auth

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

func CreateEmailMessage(code big.Int, email string) string {
	msg := fmt.Sprintf("Subject: %s\r\n", "Confirmation your email in Kanban-desk") +
    fmt.Sprintf("From: %s\r\n", cfgEmail.EmailFrom) +
    fmt.Sprintf("To: %s\r\n", email) +
    "MIME-Version: 1.0\r\n" +
    "Content-Type: text/plain; charset=\"UTF-8\"\r\n" +
    "\r\n" + // пустая строка отделяет заголовки от тела письма
    fmt.Sprintf("Your confirmation code: %v", code.String())
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
