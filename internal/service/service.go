// Package service - это пакет, который содержит логику обработки сообщений от пользователей
// и взаимодействия с внешними сервисами, такими как Salute и GigaChat.
package service

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/anatolyi0311/bot-messenger/internal/config"
	"github.com/anatolyi0311/bot-messenger/internal/models"
	"github.com/anatolyi0311/bot-messenger/internal/storage"
	"github.com/sirupsen/logrus"
)

// Auth - это структура, которая будет хранить информацию об авторизации для внешних сервисов, таких как Salute и GigaChat.
type Auth struct {
	AccessToken string
	ExpiresAt   time.Time
}

// Worker - это структура, которая будет обрабатывать сообщения от пользователей и взаимодействовать с внешними сервисами.
type Worker struct {
	cfg            *config.Config
	storage        *storage.Storage
	client         *http.Client
	authSalute     *Auth
	authGigaChat   *Auth
	MessageChannel chan models.UserMessage
}

// InitWorker - это функция, которая инициализирует новый экземпляр Worker с заданной конфигурацией и хранилищем.
func InitWorker(cfg *config.Config, storage *storage.Storage) *Worker {
	// Инициализация HTTP-клиента с настройками для работы с внешними сервисами, которые могут использовать самоподписанные сертификаты.
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	// Инициализация Worker с конфигурацией, хранилищем и HTTP-клиентом
	w := &Worker{
		cfg:            cfg,
		storage:        storage,
		client:         client,
		authSalute:     &Auth{},
		authGigaChat:   &Auth{},
		MessageChannel: make(chan models.UserMessage),
	}

	// Получение токенов для Salute и GigaChat при инициализации Worker,
	// чтобы они были готовы к использованию при обработке сообщений от пользователей.
	w.getTokenSalute()
	w.getTokenGigaChat()

	logrus.Info("Worker initialized successfully with tokens for Salute and GigaChat")

	return w
}
