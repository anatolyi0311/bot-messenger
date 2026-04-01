package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anatolyi0311/bot-messenger/bot"
	"github.com/anatolyi0311/bot-messenger/internal/config"
	"github.com/anatolyi0311/bot-messenger/internal/db"
	"github.com/anatolyi0311/bot-messenger/internal/storage"
	"github.com/anatolyi0311/bot-messenger/migrations"
	"github.com/sirupsen/logrus"
)

func main() {
	// Загрузка конфигурации
	cfg := config.LoadConfig()

	// Инициализация БД
	dbPostgres, err := db.InitPostgresDB(&cfg.Postgres)
	if err != nil {
		logrus.Fatalln(err)
	}

	// Выполнение миграций
	err = migrations.Up(dbPostgres)
	if err != nil {
		logrus.Fatalln(err)
	}
	defer func() {
		migrations.Down(dbPostgres)
		logrus.Info("Migrations down")
	}()

	// Инициализация хранилища
	store := storage.New(dbPostgres)

	// Создание контекста для управления жизненным циклом бота
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Инициализация и запуск бота
	bot := bot.New(cfg, store)
	bot.Route()

	// Запуск бота в отдельной горутине
	go func() {
		logrus.Info("Starting bot...")
		bot.Start(ctx)
	}()

	// Ожидание сигнала для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit
	logrus.Info("Shutting down gracefully...")

	cancel()

	// Даем боту время завершить обработку текущих сообщений
	shutdownTimeout := 2 * time.Second
	time.Sleep(shutdownTimeout)

	// Остановка бота
	bot.Stop()
	logrus.Info("Shutdown completed")
}
