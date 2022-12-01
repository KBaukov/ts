package main

import (
	"fmt"
	"github.com/KBaukov/ts/config"
	"github.com/KBaukov/ts/db"
	"github.com/KBaukov/ts/handle"
	"log"
	"net/http"
	"os"
)

var (
	//configurationPath = flag.String("config", "config.json", "Путь к файлу конфигурации")
	Cfg = config.LoadConfig("config.json")
)

func init() {

}

func main() {

	if Cfg.LoggerPath != "" {
		// Логер только добавляет данные
		logFile, err := os.OpenFile(Cfg.LoggerPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			log.Printf("Ошибка открытия файла лога: %v", err)
		} else {
			defer logFile.Close()
			log.SetOutput(logFile)
		}
	}

	psqlconn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		Cfg.DbConnection.DBHost,
		Cfg.DbConnection.DBPort,
		Cfg.DbConnection.DBUser,
		Cfg.DbConnection.DBPass,
		Cfg.DbConnection.DBName,
	)

	db, err := db.NewDB(psqlconn)
	if err != nil {
		log.Printf("Не удалось подключиться к базе данных: %v", err)
		return
	} else {
		db.Conn.SetMaxOpenConns(100)
		db.Conn.SetMaxIdleConns(5)
		stats := db.Conn.Stats().OpenConnections
		log.Printf("Open connections:", stats)
		_, err = db.Conn.Exec("select count(*) from event;")
		if err != nil {
			log.Printf("Не удалось установить настройки базы данных: %v", err, err)
			return
		}
		defer db.Conn.Close()
	}

	//frf := Cfg.FrontRoute.WebResFolder

	//http.HandleFunc("/logout", handle.ServeLogout(db))
	//http.HandleFunc("/login", handle.ServeLogin(db))
	http.HandleFunc("/", handle.ServeHome)
	//http.HandleFunc("", handle.ServeHome)
	//http.HandleFunc("/"+frf+"/", handle.ServeWebRes)
	http.HandleFunc("/css/", handle.ServeWebRes)
	http.HandleFunc("/js/", handle.ServeWebRes)
	http.HandleFunc("/images/", handle.ServeWebRes)
	http.HandleFunc("/paysuccess", handle.ServePagesRes)
	http.HandleFunc("/crocusrules", handle.ServePagesRes)
	http.HandleFunc("/confidential", handle.ServePagesRes)
	http.HandleFunc("/vozvrat", handle.ServePagesRes)
	http.HandleFunc("/oferta", handle.ServePagesRes)
	http.HandleFunc("/api/", handle.ServeApi(db))
	http.HandleFunc("/ws", handle.ServeWs(db))

	listenString := Cfg.Server.Address + ":" + Cfg.Server.Port
	log.Print("Сервер запущен: ", listenString)

	if Cfg.Server.TLS {
		err := http.ListenAndServeTLS(listenString, Cfg.Server.CertificatePath, Cfg.Server.KeyPath, nil)
		if err != nil {
			log.Printf("Ошибка веб-сервера: %v", err)
		}
	} else {
		err := http.ListenAndServe(listenString, nil)
		if err != nil {
			log.Printf("Ошибка веб-сервера: %v", err)
		}
	}

	log.Print("Сервер запущен: ", listenString)
}

//func inBackground(db db.Database) {
//	ticker := time.NewTicker(30 * time.Second) //time.Minute)
//
//	for now := range ticker.C {
//		db.ClearExpiredReserves()
//		log.Println(now, "#### Clear expired reserved success #####")
//	}
//}
