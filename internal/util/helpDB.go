package util

import (
	"database/sql"

	"golang.org/x/crypto/bcrypt"
)

func IsNullStringDb(str string) sql.NullString {
	var res sql.NullString
	if str == "" {
		res = sql.NullString{String: "", Valid: false}
	} else {
		res = sql.NullString{String: str, Valid: true}
	}
	return res
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err 
}

func CheckPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}