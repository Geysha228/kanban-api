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

	//регистрация
	mux.Handle("/user/reg", LoggerMiddleware(MethodCheckMiddleware(http.MethodPost, RegisterHandler(userRepo))))

	//подтверждение почты
	mux.Handle("/user/reg/confirm-email", LoggerMiddleware(MethodCheckMiddleware(http.MethodPost, ConfirmEmailHandler(userRepo))))
	
	//отправка нового кода подтверждения пользователя
	mux.Handle("/user/reg/confirm-email/new-code", LoggerMiddleware(MethodCheckMiddleware(http.MethodPost, SendNewConfirmationCodeHandler(userRepo))))


	
	return mux
}