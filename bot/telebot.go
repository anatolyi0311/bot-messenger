// Package bot - это пакет, который содержит логику создания и управления телеграм-ботом,
package bot

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
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

func (b *Bot) Route() {
	b.bot.Handle("/start", b.startHandler)
	b.bot.Handle("/get", b.getHandler)
	b.bot.Handle(tb.OnText, b.textHandler)
	b.bot.Handle(tb.OnVoice, b.voiceHandler)
	logrus.Info("Handlers registered")
}

// Start - это метод, который запускает бота и обрабатывает входящие сообщения от пользователей.
// Он также слушает канал для получения сообщений от сервиса и отправляет их пользователям.
func (b *Bot) Start(ctx context.Context) {
	// Запуск горутины для обработки сообщений от сервиса и отправки их пользователям.
	go func() {
		for {
			select {
			case <-ctx.Done():
				// Получение сигнала завершения работы и корректное завершение горутины для обработки сообщений от сервиса.
				logrus.Info("TeleBot: received shutdown signal, stopping...")
				return
			case m := <-b.worker.MessageChannel:
				// Получение сообщения от сервиса и отправка его пользователю через телеграм-бота.
				logrus.Infof("TeleBot: sending message to chat ID %d: %s", m.ChatID, m.Message)
				b.bot.Send(&tb.Chat{ID: m.ChatID}, m.Message)
			}
		}
	}()

	// Запуск сервиса Worker для обработки сообщений от пользователей и взаимодействия с внешними сервисами.
	go b.RunWorker(ctx)

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

// RunWorker - это метод, который запускает сервис Worker для обработки сообщений от пользователей и взаимодействия с внешними сервисами.
// Он принимает контекст для управления жизненным циклом Worker и обеспечивает его корректное завершение при получении сигнала остановки.
func (b *Bot) RunWorker(ctx context.Context) {
	logrus.Info("Starting Worker...")
	go b.worker.Run(ctx)
}

func (b *Bot) startHandler(c tb.Context) error {
	return b.textHandler(c)
}

func (b *Bot) textHandler(c tb.Context) error {
	if c.Text() == "/start" {
		id, err := b.worker.SaveIncomingID(c.Chat().ID)
		if err != nil {
			logrus.Error(err)
			return c.Send("Error saving chat ID")
		}
		logrus.Infof("Chat ID saved successfully with ID: %d", id)
		return c.Send(fmt.Sprintf("Chat ID saved successfully with ID: %d", id))
	}
	return c.Send("Your text is not defined:" + c.Text())
}

// voiceHandler - это метод, который обрабатывает входящие голосовые сообщения от пользователей,
// сохраняет их в базе данных и отправляет подтверждение пользователю.
func (b *Bot) voiceHandler(c tb.Context) error {
	// Получение голосового сообщения от пользователя и сохранение его в базе данных с помощью сервиса Worker.
	msg := c.Message().Voice
	// Получение файла голосового сообщения от Telegram API с помощью метода File() и сохранение его в переменной file.
	// Если при получении файла возникает ошибка, то пользователю отправляется сообщение об ошибке,
	// а также логируется ошибка для дальнейшего анализа и устранения проблемы.
	file, err := b.bot.File(&tb.File{FileID: msg.FileID})
	if err != nil {
		logrus.Error(err)
		return c.Send("Error retrieving voice file")
	}
	// Чтение данных голосового сообщения из файла, полученного от Telegram API, и сохранение его в базе данных с помощью сервиса Worker.
	// Если при чтении файла или сохранении данных возникает ошибка, то пользователю отправляется сообщение об ошибке,
	// а также логируется ошибка для дальнейшего анализа и устранения проблемы.
	voiceBytes, err := io.ReadAll(file)
	if err != nil {
		logrus.Error(err)
		return c.Send("Error reading voice file")
	}
	// Сохранение данных голосового сообщения в базе данных с помощью сервиса Worker.
	// Если при сохранении данных возникает ошибка, то пользователю отправляется сообщение об ошибке,
	// а также логируется ошибка для дальнейшего анализа и устранения проблемы.
	id, err := b.worker.SaveIncomingVoice(voiceBytes, c.Chat().ID)
	if err != nil {
		logrus.Error(err)
		return c.Send("Error saving voice file")
	}
	logrus.Infof("Voice file saved successfully with ID: %d", id)
	// Отправка пользователю сообщения об успешном сохранении голосового сообщения, включая его идентификатор (voiceID) для дальнейшей обработки.
	return c.Send(fmt.Sprintf("Voice file saved successfully with ID: %d", id))
}

// getHandler - это метод, который обрабатывает входящие текстовые сообщения от пользователей с командой /get,
func (b *Bot) getHandler(c tb.Context) error {
	args := c.Args()
	if len(args) == 0 {
		return c.Send("Пожалуйста, укажите ID встречи для получения краткого содержания. Например: /get 123")
	}
	voiceID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return c.Send("Неверный формат ID встречи. Пожалуйста, укажите числовой ID. Например: /get 123")
	}
	// Получение краткого содержания для текста распознанной речи с помощью сервиса Worker, используя идентификатор голосового сообщения (voiceID) и идентификатор чата (chatID) для получения краткого содержания для данного голосового сообщения. Если при получении краткого содержания произошла ошибка, то возвращаем сообщение об ошибке, иначе возвращаем краткое содержание для данного голосового сообщения.
	summary, err := b.worker.GetSummaryByID(voiceID, c.Chat().ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Send("Встреча не найдена")
		}
		return c.Send("Произошла внутренняя ошибка сервера")
	}
	// Отправка пользователю краткого содержания для данного голосового сообщения, полученного с помощью сервиса Worker, включая идентификатор голосового сообщения (voiceID) для дальнейшей обработки.
	return c.Send(summary)
}
