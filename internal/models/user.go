package models

import "time"

type User struct {
	ID                              int    `json:"id"`
	Login                           string `json:"login" validate:"required,min=6,max=20"`
	FirstName                       string `json:"first_name" validate:"required,min=2,max=25"`
	LastName                        string `json:"last_name" validate:"required,min=2,max=25"`
	Patronymic                      string `json:"patronymic" validate:"omitempty,min=2,max=25"`
	Password                        string `json:"password" validate:"required,min=8,max=20,alphanum"`
	Position                        string `json:"position" validate:"omitempty,min=2,max=50"`
	Email                           string `json:"email" validate:"required,min=4,max=30,email"`
	IsConfirmed                     bool   `json:"is_confirmed"`
	UserCode User_Codes `json:"user_codes"`
}

type UserConfirm struct {
	Email                 string `json:"email" validate:"required,min=4,max=30,email"`
	EmailConfirmationCode string `json:"email_confirmation_code" validate:"required,len=6"`
}

type UserAutho struct {
	LoginEmail string `json:"login_email" validate:"required,min=4,max=30"`
	Password string `json:"password" validate:"required,min=8,max=20,alphanum"`
}

type UserOnlyLoginEmail struct{
	LoginEmail string `json:"login_email" validate:"required,min=4,max=30"`
}

type User_Codes struct {
	EmailConfirmationCode           string `json:"email_confirmation_code" validate:"required,len=6"`
	ExpirationEmailConfirmationCode time.Time `json:"expiration_email_confirmation_code" validate:"required"`
	EmailConfirmationPasswordCode           string `json:"email_confirmation_password_code" validate:"omitempty,len=6"`
	ExpirationEmailConfirmationPasswordCode time.Time `json:"expiration_email_confirmation_password_code"`
}

type UserResetPassword struct{
	Email                 string `json:"email" validate:"required,min=4,max=30,email"`
	Password string `json:"password" validate:"required,min=8,max=20,alphanum"`
	EmailConfirmationCode           string `json:"email_confirmation_code" validate:"omitempty,len=6"`
}

