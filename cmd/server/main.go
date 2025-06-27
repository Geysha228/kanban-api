package main

import (
	"fmt"
	"kanban-api/config"
	"kanban-api/internal/api"
	"kanban-api/internal/db"
	"kanban-api/internal/util"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

var server *http.Server

func init() {

	//настройка логирования
	go util.StartLogWriter()

	//загрузка конфига из файла
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Can't load config: %v", err)
	}

	//настройка конфига для сообщений
	util.SetConfigEmail()
	
	//подключение к БД
	if err = db.ConnectDB(config); err != nil {
		log.Fatalf("Can't connect to DB: %v", err)
	} else {
		util.LogWrite(fmt.Sprintf("Succesful connect to database: %s", config.HTTPServer.Address))
	}

	//проверка подключения к БД
	if err = db.DBPing(); err != nil {
		log.Fatalf("Can't ping DB: %v", err)
	} else {
		util.LogWrite(fmt.Sprintf("Succesful ping database: %s", config.HTTPServer.Address))
	}

	server = &http.Server{
		Addr: config.HTTPServer.Address,
		Handler: api.SetupRouter(),
	}
}

func main(){
	defer db.CloseDB()
	defer util.CloseLogChan()
	
	//Доделать обработку ошибки при запуске сервера
	err := server.ListenAndServe()
	if err != nil{
		util.LogWrite(fmt.Sprintf("Can't start server: %v", err))
	}
}