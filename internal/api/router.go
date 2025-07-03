package api

import (
	"kanban-api/internal/db"
	"kanban-api/internal/repository"
	"net/http"
)

//функция возвращения нового mux
//со всеми маршрутами
func SetupRouter() http.Handler{
	userRepo := repository.NewUserRepository(db.GetDB())
	
	mux := http.NewServeMux()


	//РЕГИСТРАЦИЯ

	//регистрация
	mux.Handle("/user/reg", CORSMiddleware(LoggerMiddleware(MethodCheckMiddleware(http.MethodPost, RegisterHandler(userRepo)))))

	//подтверждение почты
	mux.Handle("/user/reg/confirm-email", CORSMiddleware(LoggerMiddleware(MethodCheckMiddleware(http.MethodPost, ConfirmEmailHandler(userRepo)))))
	
	//отправка нового кода подтверждения пользователя
	mux.Handle("/user/reg/confirm-email/new-code", CORSMiddleware(LoggerMiddleware(MethodCheckMiddleware(http.MethodPost, SendNewConfirmationCodeHandler(userRepo)))))



	//АВТОРИЗАЦИЯ

	//отправка кода для сброса пароля
	mux.Handle("/user/autho/forgot-password", CORSMiddleware(LoggerMiddleware(MethodCheckMiddleware(http.MethodPost, SendNewConfirmationPasswordCodeHandler(userRepo)))))
	
	//проверка кода для сброса пароля
	mux.Handle("/user/autho/forgot-password/code", CORSMiddleware(LoggerMiddleware(MethodCheckMiddleware(http.MethodPost, CheckConfirmationPasswordCodeHandler(userRepo)))))

	//сброс пароля
	mux.Handle("/user/autho/forgot-password/reset", CORSMiddleware(LoggerMiddleware(MethodCheckMiddleware(http.MethodPost, ResetPasswordHandler(userRepo)))))

	//авторизация
	mux.Handle("/user/autho", CORSMiddleware(LoggerMiddleware(MethodCheckMiddleware(http.MethodPost, AuthorizationUserHandler(userRepo)))))
	


	//ПРОФИЛЬ ПОЛЬЗОВАТЕЛЯ

	//изменение данных пользователя
	mux.Handle("/user/profile/change-info", CORSMiddleware(LoggerMiddleware(MethodCheckMiddleware(http.MethodPatch, ChangeUserInfoHandler(userRepo)))))
	
	//получение данных пользователя о себе
	mux.Handle("/user/profile", CORSMiddleware(LoggerMiddleware(MethodCheckMiddleware(http.MethodGet, GetInfoAboutUser(userRepo)))))
	return mux
}