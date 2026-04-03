package config

import (
	"encoding/json"
	"os"

	"github.com/sirupsen/logrus"
)

type BotConfig struct {
	Token string `json:"token"`
}

type url struct {
	URL string `json:"url"`
}

type SaluteSpeech struct {
	ClientSecret string `json:"client_secret"`
	ClientID     string `json:"client_id"`
}

type GigaChat struct {
	URL              string `json:"devices_url"`
	AuthorizationKey string `json:"Authorization_Key"`
	ClientSecret     string `json:"client_secret"`
	ClientID         string `json:"client_id"`
}

type PostgresConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	DBName   string `json:"dbname"`
	SSLMode  string `json:"sslmode"`
	Password string `json:"password"`
}

type Config struct {
	Bot            BotConfig      `json:"bot"`
	Postgres       PostgresConfig `json:"postgres"`
	Salute         SaluteSpeech   `json:"salute_speech"`
	GigaChat       GigaChat       `json:"giga_chat"`
	IntervalTicker int            `json:"interval_ticker"`
	SberURL        string         `json:"sber_devices_url"`
	SmartSpeechURL string         `json:"smart_speech_url"`
	GigaChatURL    string         `json:"giga_chat_devices_url"`
}

func LoadConfig() *Config {
	var config Config
	data, err := os.ReadFile(os.Getenv("CONFIG"))
	if err != nil {
		logrus.Fatalf("cannot load bot config %v", err)
	}
	if err = json.Unmarshal(data, &config); err != nil {
		logrus.Fatalf("cannot load bot config %v", err)
	}
	return &config
}
