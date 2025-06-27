package api

import (
	"encoding/json"
	"fmt"
	"kanban-api/internal/models"
	"kanban-api/internal/util"
	"net/http"
)

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		util.LogWrite(fmt.Sprintf("Started %s %s", r.Method, r.URL.Path))
		next.ServeHTTP(w, r)
		util.LogWrite(fmt.Sprintf("Completed %s %s", r.Method, r.URL.Path))
	})
}

func MethodCheckMiddleware(allowedMethod string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		if r.Method != allowedMethod {
			util.LogWrite(fmt.Sprintf("Not allowed method: %s", r.Method))
			w.Header().Set("Content-Type", "application/json")
			errors := []models.APIError{
				    {Error: "Not allowed method", ErrorCode: "2301"},
			}
			w.WriteHeader(http.StatusMethodNotAllowed)
			response := models.ErrorResponse{
				Errors: errors,
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func CORSMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "http://26.45.225.141:3000") 
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        next.ServeHTTP(w, r)
    })
}