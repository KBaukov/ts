package config

import (
	"encoding/json"
	"log"
	"os"
)

type Configuration struct {
	DbConnection struct {
		DBHost string
		DBPort int
		DBName string
		DBUser string
		DBPass string
	}
	LoggerPath         string
	Server             struct {
		Address         string
		Port            string
		TLS             bool
		CertificatePath string
		KeyPath         string
	}
	WsConfig struct {
		WsAllowedOrigin string
		BrPref string
		PingPeriod int
		WriteWait int
		MaxMessageSize int64
		PingWait int
		PongWait int
		CloseGracePeriod int
	}
	FrontRoute struct {
		WebResFolder string
		MainTemplate string
		LoginTemplate string
	}
}

// loadConfig читает и парсит настройки сервиса
func LoadConfig(configPath string) Configuration {
	var config Configuration
	file, err := os.Open(configPath)
	if err != nil {
		log.Fatalf("Не удалось открыть файл конфигурации: %v", err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalf("Не удалось прочесть файл конфигурации: %v", err)
	}
	return config
}
