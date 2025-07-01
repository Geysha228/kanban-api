package repository

import (
	"database/sql"
	"fmt"
	"kanban-api/internal/models"
	"kanban-api/internal/util"
	"time"

	_ "github.com/lib/pq"
)

type UsRepo interface{
	GetUserById(id int) models.User
	CheckUserByEmail(email string) (bool, error)
	CheckUserByLogin(login string) (bool, error)
	CreateUser(user *models.User) error
	GetHashPasswordFromDb(login string) (password string, err error)
	ConfirmEmail(user models.UserConfirm) (bool, error)
	GetHashPasswordAndIDAndEmailFromDB(loginEmail string) (user models.UserAutho, err error)
	CreateNewConfirmationEmailCode(user models.UserConfirm) (err error)
	GetEmailAndIDByLoginOrEmail(loginEmail string) (user models.UserConfirm,err error)
	CreateNewConfirmationEmailPasswordCode(user models.UserConfirm) (err error)
	CreateNewPassword(user models.UserResetPassword) (result bool, err error)
	CheckConfirmationEmailPasswordCode(user models.UserConfirm) (bool,error)
}

type UserRepository struct {
    db *sql.DB
}


func NewUserRepository(db *sql.DB) *UserRepository {
    return &UserRepository{db: db}
}

func (usRepo *UserRepository)GetUserById(id int) models.User {
	var user models.User

	err := usRepo.db.QueryRow("SELECT id, login, first_name, last_name, patronymic, password, \"position\", email FROM public.\"User\" WHERE id = $1", id).Scan(&user.ID, &user.Login, &user.FirstName, &user.LastName, &user.Patronymic, &user.Password, &user.Position, &user.Email)
	if err != nil{
		util.LogWrite(fmt.Sprintf("%v", err))
		return models.User{}
	}
	return user
}

func (usRepo *UserRepository) CheckUserByEmail(email string) (bool, error) {
	var id int
	err := usRepo.db.QueryRow("SELECT id from public.\"User\" WHERE email = $1", email).Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		util.LogWrite(fmt.Sprintf("%v", err))
		return false, err
	}
	return true, nil
}

func (usRepo *UserRepository) CheckUserByLogin(login string) (bool, error) {
	var id int
	err := usRepo.db.QueryRow("SELECT id from public.\"User\" WHERE login = $1", login).Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		util.LogWrite(fmt.Sprintf("%v", err))
		return false, err
	}
	return true, nil
}

func (usRepo *UserRepository) CreateUser(user *models.User) error {
	queryUser := `INSERT INTO public."User" (login, first_name, last_name, patronymic, password, "position", email, "is_сonfirmed") 
                  VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id;`
	var userID int
    err := usRepo.db.QueryRow(queryUser, user.Login, user.FirstName, user.LastName, util.IsNullStringDb(user.Patronymic),
        user.Password, util.IsNullStringDb(user.Position), user.Email, user.IsConfirmed).Scan(&userID)
	if err != nil {
        return err
    }
	queryCode := `INSERT INTO public."User_Codes" (user_id, email_сonfirmation_сode, expiration_email_confirmation_code) VALUES ($1, $2, $3);`
	_, err = usRepo.db.Exec(queryCode, userID, user.UserCode.EmailConfirmationCode, user.UserCode.ExpirationEmailConfirmationCode)
    if err != nil {
        return err
    }
	return nil
}

func (usRepo *UserRepository) GetHashPasswordFromDb(login string) (password string, err error) {
	query := `SELECT password FROM public."User" WHERE login = $1 OR email = $1`
	err = usRepo.db.QueryRow(query, login).Scan(&password)
	return password, err
}

func (usRepo *UserRepository) ConfirmEmail(user models.UserConfirm) (bool, error) {
	queryBool := `
		SELECT u.id
		FROM public."User" u
		JOIN public."User_Codes" uc ON uc.user_id = u.id
		WHERE u.email = $1
		AND uc.expiration_email_confirmation_code > NOW()
		AND uc."email_сonfirmation_сode" = $2`
	var userID int
	err := usRepo.db.QueryRow(queryBool, user.Email, user.EmailConfirmationCode).Scan(&userID)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	_, err = usRepo.db.Exec(`UPDATE public."User" SET is_сonfirmed = true WHERE id = $1`, userID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (usRepo *UserRepository) GetHashPasswordAndIDAndEmailFromDB(loginEmail string) (user models.UserAutho, err error){
	query := `SELECT password, id, email FROM public."User" WHERE login = $1 OR email = $1`
	err = usRepo.db.QueryRow(query, loginEmail).Scan(&user.Password, &user.ID, &user.LoginEmail)
	return user, err
}

func (usRepo *UserRepository) CreateNewConfirmationEmailCode(user models.UserConfirm) (err error){
	query := `UPDATE public."User_Codes" SET "email_сonfirmation_сode" = $1, expiration_email_confirmation_code = $2 WHERE user_id = $3`
	_, err = usRepo.db.Exec(query, user.EmailConfirmationCode, time.Now().Add(15 * time.Minute) ,user.ID)
	return err
}

func (usRepo *UserRepository) GetEmailAndIDByLoginOrEmail(loginEmail string) (user models.UserConfirm,err error){
	query := `SELECT email, id FROM public."User" WHERE login = $1 OR email = $1`
	err = usRepo.db.QueryRow(query, loginEmail).Scan(&user.Email, &user.ID)
	return user, err
}

func (usRepo *UserRepository) CreateNewConfirmationEmailPasswordCode(user models.UserConfirm) (err error){
	query := `UPDATE public."User_Codes" SET email_confirmation_password_code = $1, expiration_email_confirmation_password_code = $2 WHERE user_id = $3`
	_, err = usRepo.db.Exec(query, user.EmailConfirmationCode, time.Now().Add(15 * time.Minute) ,user.ID)
	return err
}

func (usRepo *UserRepository) CheckConfirmationEmailPasswordCode(user models.UserConfirm) (bool,error){
	query := `SELECT u.id FROM public."User" u
	JOIN "User_Codes" uc ON u.id = uc.user_id 
	WHERE uc.email_confirmation_password_code = $1
	AND u.email = $2
	AND uc.expiration_email_confirmation_password_code > NOW()`
	var id int
	err := usRepo.db.QueryRow(query, user.EmailConfirmationCode, user.Email).Scan(&id)
	if err != nil {
		return false, err
	}
	if id == 0 {
		return false, nil
	}
	return true, nil
}

func (usRepo *UserRepository) CreateNewPassword(user models.UserResetPassword) (result bool, err error){
	query := `UPDATE "User" u
		SET password = $1
		FROM "User_Codes" uc
		WHERE u.id = uc.user_id
		  AND u.email = $2
		  AND uc.email_confirmation_password_code = $3
		  AND uc.expiration_email_confirmation_password_code > NOW();`
	res, err := usRepo.db.Exec(query, user.Password, user.Email, user.EmailConfirmationPasswordCode)
	if err != nil {
		return false, err
	}
	rowsafected, err := res.RowsAffected()
	if err != nil {
		return false, err 
	}
	if rowsafected == 0{
		return false, nil
	}
	query = `UPDATE "User_Codes" uc
	SET expiration_email_confirmation_password_code = NOW()
	FROM "User" u
	WHERE u.id = uc.user_id
		AND u.email = $1`

	res, err = usRepo.db.Exec(query, user.Email)
	if err != nil {
		return false, err
	}
	rowsafected, err = res.RowsAffected()
	if err != nil {
		return false, err 
	}
	if rowsafected == 0{
		return false, nil
	}
	return true, nil
}