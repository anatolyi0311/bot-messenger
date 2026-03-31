package main

import (
	"github.com/anatolyi0311/bot-messenger/bot"
	"github.com/anatolyi0311/bot-messenger/internal/config"
	"github.com/anatolyi0311/bot-messenger/internal/db"
	"github.com/anatolyi0311/bot-messenger/migrations"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg := config.LoadConfig()

	// Инициализация БД
	dbClient, err := db.InitPostgresDB(&cfg.Postgres)
	if err != nil {
		logrus.Fatalln(err)
	}

	err = migrations.Up(dbClient)
	if err != nil {
		logrus.Fatalln(err)
	}
	defer func() {
		migrations.Down(dbClient)
		logrus.Info("Migrations down")
	}()

	bot.Bot(cfg)
}
