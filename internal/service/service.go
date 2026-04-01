// Package service - это пакет, который содержит логику обработки сообщений от пользователей
// и взаимодействия с внешними сервисами, такими как Salute и GigaChat.
package service

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
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

// Run - это метод, который запускает Worker и обрабатывает сообщения от пользователей, а также взаимодействует с внешними сервисами для обработки голосовых сообщений.
func (w *Worker) Run(ctx context.Context) {
	// Инициализация таймера для периодической обработки голосовых сообщений, которые готовы для загрузки в Salute для распознавания речи.
	ticker := time.NewTicker(time.Duration(w.cfg.IntervalTicker) * time.Second)
	defer ticker.Stop()
	defer close(w.MessageChannel)
	for {
		select {
		case <-ctx.Done():
			// Получение сигнала завершения работы и остановка Worker, а также закрытие канала для сообщений от пользователей.
			logrus.Info("Worker: received shutdown signal, stopping...")
			return
		case <-ticker.C:
			logrus.Info("Worker: ticker ticked, starting upload process...")
			// Выполнение функции upload() при каждом срабатывании таймера, которая обрабатывает голосовые сообщения, готовые для загрузки в Salute для распознавания речи.
			w.upload()
			logrus.Info("Worker: upload process completed, starting recognize process...")
			// Выполнение функции recognize() после завершения upload(), которая обрабатывает голосовые сообщения, готовые для распознавания речи, и взаимодействует с Salute для получения результатов распознавания.
			w.recognize()
			logrus.Info("Worker: recognize process completed, starting check status process...")
			// Выполнение функции checkStatus() после завершения recognize(), которая проверяет статус обработки голосовых сообщений в Salute и обновляет их статус в базе данных, а также отправляет пользователю сообщение об успешной обработке или ошибке при обработке голосового сообщения.
			w.checkStatus()
			logrus.Info("Worker: check status process completed")
		}
	}
}

// upload - это метод, который извлекает из базы данных список голосовых сообщений, которые находятся на этапе "BEGIN" и готовы для загрузки в Salute для распознавания речи. Для каждого такого сообщения он выполняет загрузку в Salute и обновляет статус обработки в базе данных, а также отправляет пользователю сообщение об успешной загрузке или ошибке при обработке голосового сообщения.
func (w *Worker) upload() {
	logrus.Info("Worker.upload: starting upload process...")
	uploadData, err := w.storage.GetVoiceForUpload()
	if err != nil {
		if errors.Is(err, errors.New("no voices for upload")) {
			logrus.Warn(err)
			return
		}
		logrus.Error(fmt.Errorf("Worker.upload: %w", err))
		return
	}

	for i := range uploadData {
		reqFileID, err := w.uploadExecute(uploadData[i])
		if err != nil {
			logrus.Error(err)

			go w.sendMsgFail(uploadData[i].ChatID, uploadData[i].VoiceID)

			if err = w.storage.SetStatusFail(uploadData[i].VoiceID); err != nil {
				logrus.Error(fmt.Errorf("Worker.upload: %w", err))
			}
			continue
		}
		err = w.storage.SetStatusUpload(uploadData[i].VoiceID, reqFileID)
		if err != nil {
			logrus.Error(err)
		} else {
			logrus.Infof("success upload: %d, %d", uploadData[i].VoiceID, uploadData[i].ChatID)
		}
	}
}

// getTokenGigaChat - это метод, который получает токен доступа для API GigaChat и сохраняет его в структуре Auth для дальнейшего использования при взаимодействии с API GigaChat.
func (w *Worker) sendMsgSuccess(chatID, voiceID int64) {
	w.MessageChannel <- models.UserMessage{
		ChatID:  chatID,
		Message: fmt.Sprintf("The meeting was saved successfully! You can get it by ID: %d.", voiceID),
	}
}

// sendMsgFail - это метод, который отправляет пользователю сообщение об ошибке при обработке голосового сообщения, указывая идентификатор голосового сообщения (voiceID) для повторной отправки.
func (w *Worker) sendMsgFail(chatID, voiceID int64) {
	w.MessageChannel <- models.UserMessage{
		ChatID:  chatID,
		Message: fmt.Sprintf(string("Couldn't process the appointment. ID: %d. Repeat sending."), voiceID),
	}
}

func (w *Worker) recognize() {
	recognizeData, err := w.storage.GetVoiceForRecognize()
	if err != nil {
		logrus.Error(fmt.Errorf("Worker.recognize(): %w", err))
		return
	}

	for i := range recognizeData {
		recognizeID, err := w.recognizeExecute(recognizeData[i])
		if err != nil {
			logrus.Error(err)

			go w.sendMsgFail(recognizeData[i].ChatID, recognizeData[i].VoiceID)

			if err = w.storage.SetStatusFail(recognizeData[i].VoiceID); err != nil {
				logrus.Error(fmt.Errorf("Worker.recognize(): %w", err))
			}
			continue
		}
		err = w.storage.SetStatusRecognition(recognizeData[i].VoiceID, recognizeID)
		if err != nil {
			logrus.Error(err)
		} else {
			logrus.Infof("success recognize: %d", recognizeData[i].VoiceID)
		}
	}
}

func (w *Worker) checkStatus() {
	checkStatusData, err := w.storage.GetVoiceForCheckStatus()
	if err != nil {
		logrus.Error(fmt.Errorf("Worker.checkStatus(): %w", err))
		return
	}

	for i := range checkStatusData {
		respFileID, err := w.checkStatusExecute(checkStatusData[i])
		if err != nil {
			if errors.Is(err, errors.New("status not DONE yet")) {
				logrus.Infof("status not DONE yet: %d", checkStatusData[i].VoiceID)
				continue
			}

			logrus.Error(err)

			go w.sendMsgFail(checkStatusData[i].ChatID, checkStatusData[i].VoiceID)

			if err = w.storage.SetStatusFail(checkStatusData[i].VoiceID); err != nil {
				logrus.Error(fmt.Errorf("Worker.checkStatus(): %w", err))
			}
			continue
		}
		err = w.storage.SetStatusWait(checkStatusData[i].VoiceID, respFileID)
		if err != nil {
			logrus.Error(err)
		} else {
			logrus.Infof("success check status: %d", checkStatusData[i].VoiceID)
		}
	}
}
