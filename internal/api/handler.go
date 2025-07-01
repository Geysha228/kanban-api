package api

import (
	"encoding/json"
	"fmt"
	"kanban-api/internal/models"
	"kanban-api/internal/repository"
	"kanban-api/internal/util"
	"net/http"
	"time"

	"github.com/go-playground/validator"
)

type EmailResponse struct {
	Email string `json:"email"`
}

func RegisterHandler(repo repository.UsRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		
		//Получение пользователя из JSON
		user, err := util.DecodeJSONBody[models.User](r)
        if err != nil {
			util.LogWrite("Can't parse json")
			w.Header().Set("Content-Type", "application/json")
			errors := []models.APIError{{Error: "Can't parse json", ErrorCode: "3112"},}
			w.WriteHeader(http.StatusBadRequest)
			response := models.ErrorResponse{Errors: errors,}
			json.NewEncoder(w).Encode(response)
			return
        }


		//Проверка на повторный email
		util.LogWrite(fmt.Sprintf("Send request to check email - %s\n", user.Email))
		repeatEmail, err := repo.CheckUserByEmail(user.Email)
		if err != nil {
			util.LogWrite("Bad request to DB")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			errors := []models.APIError{{Error: "Bad work DB", ErrorCode: "0121"},}
			response := models.ErrorResponse{Errors: errors,}
			json.NewEncoder(w).Encode(response)
			return
		}
		util.LogWrite(fmt.Sprintf("Response to check email - %t\n", repeatEmail))


		//Проверка на повторный логин
		util.LogWrite(fmt.Sprintf("Send request to check login - %s\n", user.Login))
		repeatLogin, err := repo.CheckUserByLogin(user.Login)
		if err != nil {
			util.LogWrite("Bad request to DB")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			errors := []models.APIError{{Error: "Bad work DB", ErrorCode: "0121"},}
			response := models.ErrorResponse{Errors: errors,}
			json.NewEncoder(w).Encode(response)
			return
		}
		util.LogWrite(fmt.Sprintf("Response to check login - %t\n", repeatLogin))	
		


		//Отправка ошибки, если имеется повторные данные
		apiErrors := make([]models.APIError, 0, 2)
		if repeatEmail {
			apiErrors = append(apiErrors, models.APIError{Error: "Repeat email", ErrorCode: "1211"})
		}
		if repeatLogin {
			apiErrors = append(apiErrors, models.APIError{Error: "Repeat login", ErrorCode: "1212"})
		}
		if len(apiErrors) > 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			response := models.ErrorResponse{Errors: apiErrors,}
			json.NewEncoder(w).Encode(response)
			return
		}



		//Создание кода для подтверждения
		number, err := util.CreateEmailCode()
		if err != nil{
			util.LogWrite(fmt.Sprintf("Can't create code for email: %v", err))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			errors := []models.APIError{{Error: "Can't create code for email", ErrorCode: "5301"},}
			response := models.ErrorResponse{Errors: errors,}
			json.NewEncoder(w).Encode(response)
			return
		}
		user.UserCode.EmailConfirmationCode = number.String()
		user.IsConfirmed = false
		user.UserCode.ExpirationEmailConfirmationCode = time.Now().Add(15 * time.Minute)



		//Валидация данных
		validate := validator.New()
		err = validate.Struct(user)
		if err != nil {
			util.LogWrite(fmt.Sprintf("Field validation error: %v", err))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)
			errors := []models.APIError{{Error: "Field validation error", ErrorCode: "0912"},}
			response := models.ErrorResponse{Errors: errors,}
			json.NewEncoder(w).Encode(response)
			return
		}

		//Хеширование пароля
		user.Password, err = util.HashPassword(user.Password)
		if err != nil {
			util.LogWrite("Can't hash password")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			errors := []models.APIError{{Error: "Can't hash password", ErrorCode: "1110"},}
			response := models.ErrorResponse{Errors: errors,}
			json.NewEncoder(w).Encode(response)
			return
		}


		//Отправление SQL-запроса на создание нового пользователя
		err = repo.CreateUser(&user)
		if err != nil{
			util.LogWrite(fmt.Sprintf("Bad request to DB: %v", err))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			errors := []models.APIError{{Error: "Bad work DB", ErrorCode: "0121"},}
			response := models.ErrorResponse{Errors: errors,}
			json.NewEncoder(w).Encode(response)
			return
		}
		util.LogWrite(fmt.Sprintf("Succesful create user: %s", user.Login))

		

		//Отправление сообщение для подтверждения почты
		msg := util.CreateEmailMessage(*number, user.Email, "This code need to confirm your profile")
		err = util.SendMail(msg, user.Email)
		if err != nil {
			util.LogWrite(fmt.Sprintf("Can't send message to email %v", err))
			w.Header().Set("Content-Type", "application/json")
			errors := []models.APIError{{Error: "Can't send message to email", ErrorCode: "9856"},}
			w.WriteHeader(http.StatusBadRequest)
			response := models.ErrorResponse{Errors: errors,}
			json.NewEncoder(w).Encode(response)
			return
		}


		util.LogWrite(fmt.Sprintf("Send confirm message to email %s", user.Email))
		w.WriteHeader(http.StatusCreated)
	}
}

func ConfirmEmailHandler(repo repository.UsRepo) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request) {

		//получение json-данных
		user, err := util.DecodeJSONBody[models.UserConfirm](r)
        if err != nil {
			util.LogWrite(fmt.Sprintf("Can't parse json: %v", err))
			w.Header().Set("Content-Type", "application/json")
			errors := []models.APIError{{Error: "Can't parse json", ErrorCode: "3112"},}
			w.WriteHeader(http.StatusBadRequest)
			response := models.ErrorResponse{Errors: errors,}
			json.NewEncoder(w).Encode(response)
			return
        }


		//Валидация данных
		validate := validator.New()
		err = validate.Struct(user)
		if err != nil {
			util.LogWrite(fmt.Sprintf("%v", err))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)
			errors := []models.APIError{{Error: "Field validation error", ErrorCode: "0912"},}
			response := models.ErrorResponse{Errors: errors,}
			json.NewEncoder(w).Encode(response)
			return
		}


		//Отправка SQL-запроса на подтверждение почты
		res, err := repo.ConfirmEmail(user)
		if err != nil{
			util.LogWrite(fmt.Sprintf("Bad request to DB: %v", err))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			errors := []models.APIError{{Error: "Bad work DB", ErrorCode: "0121"},}
			response := models.ErrorResponse{Errors: errors,}
			json.NewEncoder(w).Encode(response)
			return
		}
		if !res {
			util.LogWrite("Can't confrim email, no rows in DB")
			w.Header().Set("Content-Type", "application/json")
			errors := []models.APIError{{Error: "Can't confirm email", ErrorCode: "3813"},}
			w.WriteHeader(http.StatusBadRequest)
			response := models.ErrorResponse{Errors: errors,}
			json.NewEncoder(w).Encode(response)
			return
		}
		util.LogWrite(fmt.Sprintf("Succesful confirmation user: %s", user.Email))
		w.WriteHeader(http.StatusOK)
	}
}

func SendNewConfirmationCodeHandler(repo repository.UsRepo)http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request) {

		//получение json-данных
		user, err := util.DecodeJSONBody[models.UserAutho](r)
        if err != nil {
			util.LogWrite(fmt.Sprintf("Can't parse json: %v", err))
			w.Header().Set("Content-Type", "application/json")
			errors := []models.APIError{{Error: "Can't parse json", ErrorCode: "3112"},}
			w.WriteHeader(http.StatusBadRequest)
			response := models.ErrorResponse{Errors: errors,}
			json.NewEncoder(w).Encode(response)
			return
        }


		//Валидация данных
		validate := validator.New()
		err = validate.Struct(user)
		if err != nil {
			util.LogWrite(fmt.Sprintf("Field validation error: %v", err))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)
			errors := []models.APIError{{Error: "Field validation error", ErrorCode: "0912"},}
			response := models.ErrorResponse{Errors: errors,}
			json.NewEncoder(w).Encode(response)
			return
		}

		//Получение hash-пароля из БД и почты
		userAutho, err := repo.GetHashPasswordAndIDAndEmailFromDB(user.LoginEmail)
		if err != nil {
			util.LogWrite(fmt.Sprintf("Bad request to DB: %v", err))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			errors := []models.APIError{{Error: "Bad work DB", ErrorCode: "0121"},}
			response := models.ErrorResponse{Errors: errors,}
			json.NewEncoder(w).Encode(response)
			return
		}

		//Проверка на неверный email или логин
		if userAutho.Password == "" {
			util.LogWrite(fmt.Sprintf("Invalid login or email: %s", user.LoginEmail))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			errors := []models.APIError{{Error: "Invalid login or email", ErrorCode: "0412"},}
			response := models.ErrorResponse{Errors: errors,}
			json.NewEncoder(w).Encode(response)
			return
		}
		
		//Проверка пароля
		if res := util.CheckPasswordHash(user.Password, userAutho.Password); !res{
			util.LogWrite("Invalid password")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			errors := []models.APIError{{Error: "Invalid password", ErrorCode: "4142"},}
			response := models.ErrorResponse{Errors: errors,}
			json.NewEncoder(w).Encode(response)
			return
		}

		//Создание нового кода для проверки
		number, err := util.CreateEmailCode()
		if err != nil{
			util.LogWrite(fmt.Sprintf("Can't create code for email: %v", err))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			errors := []models.APIError{{Error: "Can't create code for email", ErrorCode: "5301"},}
			response := models.ErrorResponse{Errors: errors,}
			json.NewEncoder(w).Encode(response)
			return
		}

		
		//Отправление в БД нового кода
		userConf := models.UserConfirm{ID: userAutho.ID, Email: userAutho.LoginEmail, EmailConfirmationCode: number.String()}
		err = repo.CreateNewConfirmationEmailCode(userConf)
		if err != nil {
			util.LogWrite(fmt.Sprintf("Bad request to DB: %v", err))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			errors := []models.APIError{{Error: "Bad work DB", ErrorCode: "0121"},}
			response := models.ErrorResponse{Errors: errors,}
			json.NewEncoder(w).Encode(response)
			return
		}
		

		//отправление кода на почту
		msg := util.CreateEmailMessage(*number, userConf.Email, "This code need to confirm your profile")
		err = util.SendMail(msg, userConf.Email)
		if err != nil {
			util.LogWrite(fmt.Sprintf("Can't send message to email %v", err))
			w.Header().Set("Content-Type", "application/json")
			errors := []models.APIError{{Error: "Can't send message to email", ErrorCode: "9856"},}
			w.WriteHeader(http.StatusBadRequest)
			response := models.ErrorResponse{Errors: errors,}
			json.NewEncoder(w).Encode(response)
			return
		}
		util.LogWrite(fmt.Sprintf("Send confirm message to email %s", userConf.Email))
		w.WriteHeader(http.StatusOK)
	}
}

func SendNewConfirmationPasswordCodeHandler(repo repository.UsRepo)http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request) {
		//получение json-данных
		user, err := util.DecodeJSONBody[models.UserOnlyLoginEmail](r)
        if err != nil {
			util.LogWrite(fmt.Sprintf("Can't parse json: %v", err))
			w.Header().Set("Content-Type", "application/json")
			errors := []models.APIError{{Error: "Can't parse json", ErrorCode: "3112"},}
			w.WriteHeader(http.StatusBadRequest)
			response := models.ErrorResponse{Errors: errors,}
			json.NewEncoder(w).Encode(response)
			return
        }

		//валидация
		validate := validator.New()
		err = validate.Struct(user)
		if err != nil {
			util.LogWrite(fmt.Sprintf("%v", err))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)
			errors := []models.APIError{{Error: "Field validation error", ErrorCode: "0912"},}
			response := models.ErrorResponse{Errors: errors,}
			json.NewEncoder(w).Encode(response)
			return
		}

		//получение почты пользователя
		userConf, err := repo.GetEmailAndIDByLoginOrEmail(user.LoginEmail)
		if err != nil{
			util.LogWrite(fmt.Sprintf("Bad request to DB: %v", err))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			errors := []models.APIError{{Error: "Bad work DB", ErrorCode: "0121"},}
			response := models.ErrorResponse{Errors: errors,}
			json.NewEncoder(w).Encode(response)
			return
		}
		if userConf.Email == ""{
			util.LogWrite("Can't find email, no rows in DB")
			w.Header().Set("Content-Type", "application/json")
			errors := []models.APIError{{Error: "Can't confrim email", ErrorCode: "3813"},}
			w.WriteHeader(http.StatusBadRequest)
			response := models.ErrorResponse{Errors: errors,}
			json.NewEncoder(w).Encode(response)
			return
		}

		//Создание кода для подтверждения пароля
		number, err := util.CreateEmailCode()
		if err != nil{
			util.LogWrite(fmt.Sprintf("Can't create code for email: %v", err))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			errors := []models.APIError{{Error: "Can't create code for email", ErrorCode: "5301"},}
			response := models.ErrorResponse{Errors: errors,}
			json.NewEncoder(w).Encode(response)
			return
		}

		userConf.EmailConfirmationCode = number.String()

		//Занесение в бд кода
		err = repo.CreateNewConfirmationEmailPasswordCode(userConf) 
		if err != nil {
			util.LogWrite(fmt.Sprintf("Bad request to DB: %v", err))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			errors := []models.APIError{{Error: "Bad work DB", ErrorCode: "0121"},}
			response := models.ErrorResponse{Errors: errors,}
			json.NewEncoder(w).Encode(response)
			return
		}

		//Отправка нового кода
		msg := util.CreateEmailMessage(*number, userConf.Email, "This code need to reset your password")
		err = util.SendMail(msg, userConf.Email)
		if err != nil {
			util.LogWrite(fmt.Sprintf("Can't send message to email %v", err))
			w.Header().Set("Content-Type", "application/json")
			errors := []models.APIError{{Error: "Can't send message to email", ErrorCode: "9856"},}
			w.WriteHeader(http.StatusBadRequest)
			response := models.ErrorResponse{Errors: errors,}
			json.NewEncoder(w).Encode(response)
			return
		}
		
		util.LogWrite(fmt.Sprintf("Send confirm password message to email %s", userConf.Email))
		resp := EmailResponse{Email: userConf.Email}
		jsonData, err := json.Marshal(resp)
		if err != nil {
			util.LogWrite("Can't parse json")
			w.Header().Set("Content-Type", "application/json")
			errors := []models.APIError{{Error: "Can't parse json", ErrorCode: "3112"},}
			w.WriteHeader(http.StatusInternalServerError)
			response := models.ErrorResponse{Errors: errors,}
			json.NewEncoder(w).Encode(response)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	}
} 

func ResetPasswordHandler(repo repository.UsRepo)http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request) {

		//Получение данных из тела запроса
		user, err := util.DecodeJSONBody[models.UserResetPassword](r)
		if err != nil {
			util.LogWrite(fmt.Sprintf("Can't parse json: %v", err))
			w.Header().Set("Content-Type", "application/json")
			errors := []models.APIError{{Error: "Can't parse json", ErrorCode: "3112"},}
			w.WriteHeader(http.StatusBadRequest)
			response := models.ErrorResponse{Errors: errors,}
			json.NewEncoder(w).Encode(response)
			return
        }
		
		//валидация запроса
		validate := validator.New()
		err = validate.Struct(user)
		if err != nil {
			util.LogWrite(fmt.Sprintf("Field validation error: %v", err))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)
			errors := []models.APIError{{Error: "Field validation error", ErrorCode: "0912"},}
			response := models.ErrorResponse{Errors: errors,}
			json.NewEncoder(w).Encode(response)
			return
		}

		//Создание нового хешированного пароля
		user.Password, err = util.HashPassword(user.Password)
		if err != nil {
			util.LogWrite("Can't hash password")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			errors := []models.APIError{{Error: "Can't hash password", ErrorCode: "1110"},}
			response := models.ErrorResponse{Errors: errors,}
			json.NewEncoder(w).Encode(response)
			return
		}

		//Отправление запроса в БД на изменение пароля
		resReq, err := repo.CreateNewPassword(user)
		if err != nil {
			util.LogWrite(fmt.Sprintf("Bad request to DB: %v", err))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			errors := []models.APIError{{Error: "Bad work DB", ErrorCode: "0121"},}
			response := models.ErrorResponse{Errors: errors,}
			json.NewEncoder(w).Encode(response)
			return
		}
		if !resReq {
			util.LogWrite("No one row updated")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			errors := []models.APIError{{Error: "No one row updated", ErrorCode: "0124"},}
			response := models.ErrorResponse{Errors: errors,}
			json.NewEncoder(w).Encode(response)
			return			
		}

		util.LogWrite(fmt.Sprintf("Succesfull update password for email: %s", user.Email))
		w.WriteHeader(http.StatusOK)
	}
}