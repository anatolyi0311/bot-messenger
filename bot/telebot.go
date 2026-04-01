// Package bot - это пакет, который содержит логику создания и управления телеграм-ботом,
package bot

import (
	"context"
	"log"
	"time"

	tb "gopkg.in/telebot.v3"

	"github.com/anatolyi0311/bot-messenger/internal/config"
	"github.com/anatolyi0311/bot-messenger/internal/service"
	"github.com/anatolyi0311/bot-messenger/internal/storage"
	"github.com/sirupsen/logrus"
)

// Bot - это структура, которая будет представлять телеграм-бота и
// содержать его настройки, а также ссылку на сервис для обработки сообщений.
type Bot struct {
	bot    *tb.Bot
	worker *service.Worker
	stop   chan struct{}
}

// New - это функция, которая инициализирует новый экземпляр Bot с заданной конфигурацией и хранилищем.
func New(cfg *config.Config, storage *storage.Storage) *Bot {
	// Настройки для телеграм-бота, включая токен и параметры опроса для получения обновлений от Telegram API.
	settings := tb.Settings{
		Token:  cfg.Bot.Token,
		Poller: &tb.LongPoller{Timeout: 3 * time.Second},
	}

	// Инициализация телеграм-бота с заданными настройками, включая токен и параметры опроса для получения обновлений от Telegram API.
	// Если инициализация прошла успешно, то возвращается новый экземпляр Bot,
	// который готов к использованию для обработки сообщений от пользователей и взаимодействия с внешними сервисами через сервис Worker.
	bot, err := tb.NewBot(settings)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	// Инициализация сервиса для обработки сообщений от пользователей и взаимодействия с внешними сервисами, такими как Salute и GigaChat.
	w := service.InitWorker(cfg, storage)
	return &Bot{
		bot:    bot,
		worker: w,
		stop:   make(chan struct{}),
	}
}

// Start - это метод, который запускает бота и обрабатывает входящие сообщения от пользователей.
// Он также слушает канал для получения сообщений от сервиса и отправляет их пользователям.
func (b *Bot) Start(ctx context.Context) {
	// Обработчик команды /start, который отправляет пользователю приветственное сообщение.
	b.bot.Handle("/start", func(c tb.Context) error {
		return c.Send("Hello world! now:" + time.Now().Format("2006-01-02 15:04:05"))
	})

	// Запуск горутины для обработки сообщений от сервиса и отправки их пользователям.
	go func() {
		for {
			select {
			case <-ctx.Done():
				logrus.Info("TeleBot: received shutdown signal, stopping...")
				return
			case m := <-b.worker.MessageChannel:
				b.bot.Send(&tb.Chat{ID: m.ChatID}, m.Message)
			}
		}
	}()

	// Запуск бота в отдельной горутине, чтобы он не блокировал основной поток выполнения.
	go func() {
		b.bot.Start()
		close(b.stop)
	}()
}

// Stop - это метод, который останавливает бота и закрывает канал для получения сообщений от сервиса.
func (b *Bot) Stop() {
	// Остановка бота и ожидание завершения его работы, чтобы гарантировать, что все ресурсы будут освобождены корректно.
	b.bot.Stop()
	<-b.stop
}
