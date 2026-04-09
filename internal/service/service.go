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
			logrus.Info("Worker: check status process completed, starting download transcription process...")
			// Выполнение функции downloadTranscription() после завершения checkStatus(), которая загружает расшифровку текста для голосовых сообщений, которые были успешно обработаны в Salute, и обновляет их статус в базе данных, а также отправляет пользователю сообщение об успешной загрузке или ошибке при загрузке расшифровки текста.
			w.downloadTranscription()
			logrus.Info("Worker: download transcription process completed, starting create summary process...")
			// Выполнение функции createSummary() после завершения downloadTranscription(), которая создает краткое содержание для голосовых сообщений, которые были успешно расшифрованы, и обновляет их статус в базе данных, а также отправляет пользователю сообщение об успешном создании краткого содержания или ошибке при создании краткого содержания.
			w.createSummary()
			logrus.Info("Worker: create summary process completed.")
		}
	}
}

// upload - это метод, который извлекает из базы данных список голосовых сообщений, которые находятся на этапе "BEGIN" и готовы для загрузки в Salute для распознавания речи. Для каждого такого сообщения он выполняет загрузку в Salute и обновляет статус обработки в базе данных, а также отправляет пользователю сообщение об успешной загрузке или ошибке при обработке голосового сообщения.
func (w *Worker) upload() {
	logrus.Info("Worker.upload: starting upload process...")
	uploadData, err := w.storage.GetVoiceForUpload()
	if err != nil {
		if errors.Is(err, fmt.Errorf("no voices for upload")) {
			logrus.Warn(err)
			return
		}
		logrus.Error(fmt.Errorf("Worker.upload: %w", err))
		return
	}
	// Проход по каждому голосовому сообщению, готовому для загрузки в Salute для распознавания речи,
	// и выполнение процесса загрузки в Salute. Если загрузка прошла успешно,
	// то обновляем статус обработки в базе данных на "UPLOAD" и сохраняем идентификатор файла
	// с результатом загрузки для данного голосового сообщения, а также отправляем пользователю сообщение
	// об успешной загрузке. Если при загрузке возникает ошибка, то отправляем пользователю сообщение
	// об ошибке и обновляем статус обработки в базе данных на "FAIL".
	for i := range uploadData {
		reqFileID, err := w.uploadExecute(uploadData[i])
		if err != nil {
			logrus.Error("Worker.upload: ", err)

			go w.sendMsgFail(uploadData[i].ChatID, uploadData[i].VoiceID)

			if err = w.storage.SetStatusFail(uploadData[i].VoiceID); err != nil {
				logrus.Error(fmt.Errorf("Worker.upload: %w", err))
			}
			continue
		}
		err = w.storage.SetStatusUpload(uploadData[i].VoiceID, reqFileID)
		if err != nil {
			logrus.Error("Worker.upload: ", err)
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

// recognize - это метод, который извлекает из базы данных список голосовых сообщений, которые находятся на этапе "UPLOAD" и готовы для распознавания речи с помощью Salute. Для каждого такого сообщения он выполняет распознавание речи с помощью Salute и обновляет статус обработки в базе данных, а также отправляет пользователю сообщение об успешной обработке или ошибке при обработке голосового сообщения.
func (w *Worker) recognize() {
	// Получение из базы данных списка голосовых сообщений, которые находятся на этапе "UPLOAD" и готовы для распознавания речи с помощью Salute. Для каждого такого сообщения он выполняет распознавание речи с помощью Salute и обновляет статус обработки в базе данных, а также отправляет пользователю сообщение об успешной обработке или ошибке при обработке голосового сообщения.
	recognizeData, err := w.storage.GetVoiceForRecognize()
	if err != nil {
		logrus.Error(fmt.Errorf("Worker.recognize(): %w", err))
		return
	}

	// Проход по каждому голосовому сообщению, готовому для распознавания речи, и выполнение процесса распознавания речи с помощью Salute.
	for i := range recognizeData {
		recognizeID, err := w.recognizeExecute(recognizeData[i])
		if err != nil {
			logrus.Error(err)

			// Отправка пользователю сообщения об ошибке при распознавании речи, указывая идентификатор голосового сообщения (voiceID) для повторной отправки.
			go w.sendMsgFail(recognizeData[i].ChatID, recognizeData[i].VoiceID)

			// Обновление статуса обработки голосового сообщения в базе данных на "FAIL",
			if err = w.storage.SetStatusFail(recognizeData[i].VoiceID); err != nil {
				logrus.Error(fmt.Errorf("Worker.recognize(): %w", err))
			}
			continue
		}
		// Обновление статуса обработки голосового сообщения в базе данных на "RECOGNITION" и сохранение идентификатора распознавания для данного голосового сообщения.
		err = w.storage.SetStatusRecognition(recognizeData[i].VoiceID, recognizeID)
		if err != nil {
			logrus.Error(err)
		} else {
			logrus.Infof("success recognize: %d", recognizeData[i].VoiceID)
		}
	}
}

// checkStatus - это метод, который извлекает из базы данных список голосовых сообщений, которые находятся на этапе "RECOGNITION" и готовы для проверки статуса обработки в Salute. Для каждого такого сообщения он выполняет проверку статуса обработки в Salute и обновляет его статус в базе данных, а также отправляет пользователю сообщение об успешной обработке или ошибке при обработке голосового сообщения.
func (w *Worker) checkStatus() {
	// Получение из базы данных списка голосовых сообщений, которые находятся на этапе "RECOGNITION" и готовы для проверки статуса обработки в Salute. Для каждого такого сообщения он выполняет проверку статуса обработки в Salute и обновляет статус в базе данных, а также отправляет пользователю сообщение об успешной обработке или ошибке при обработке голосового сообщения.
	checkStatusData, err := w.storage.GetVoiceForCheckStatus()
	if err != nil {
		logrus.Error(fmt.Errorf("Worker.checkStatus: %w", err))
		return
	}

	// Проход по каждому голосовому сообщению, готовому для проверки статуса обработки в Salute, и выполнение процесса проверки статуса обработки в Salute.
	for i := range checkStatusData {
		// Получение статуса обработки голосового сообщения в Salute с помощью API Salute. Если статус обработки еще не "DONE", то продолжаем ожидать, иначе обновляем статус в базе данных и отправляем пользователю сообщение об успешной обработке или ошибке при обработке голосового сообщения.
		respFileID, err := w.checkStatusExecute(checkStatusData[i])
		if err != nil {
			// Если статус обработки еще не "DONE", то продолжаем ожидать, иначе отправляем пользователю сообщение об ошибке при обработке голосового сообщения и обновляем статус в базе данных на "FAIL".
			if errors.Is(err, fmt.Errorf("meeting processed yet")) {
				// logrus.Infof("status not DONE yet: %d", checkStatusData[i].VoiceID)
				logrus.Warn(fmt.Errorf("status not DONE yet: %d", checkStatusData[i].VoiceID))
				continue
			}

			logrus.Error("Worker.checkStatus: ", err)

			// Отправка пользователю сообщения об ошибке при обработке голосового сообщения, указывая идентификатор голосового сообщения (voiceID) для повторной отправки.
			go w.sendMsgFail(checkStatusData[i].ChatID, checkStatusData[i].VoiceID)

			// Обновление статуса обработки голосового сообщения в базе данных на "FAIL",
			if err = w.storage.SetStatusFail(checkStatusData[i].VoiceID); err != nil {
				logrus.Error(fmt.Errorf("Worker.checkStatus: %w", err))
			}
			continue
		}
		// Обновление статуса обработки голосового сообщения в базе данных на "WAIT" и сохранение идентификатора файла с результатом распознавания для данного голосового сообщения.
		err = w.storage.SetStatusWait(checkStatusData[i].VoiceID, respFileID)
		if err != nil {
			logrus.Error(err)
		} else {
			logrus.Infof("success check status: %d", checkStatusData[i].VoiceID)
		}
	}
}

// downloadTranscription - это метод, который извлекает из базы данных список голосовых сообщений,
// которые находятся на этапе "WAIT" и готовы для загрузки расшифровки текста для голосового сообщения,
// которое было успешно обработано в Salute.
// Для каждого такого сообщения он выполняет загрузку расшифровки текста с помощью API Salute
// и обновляет статус в базе данных, а также отправляет пользователю сообщение об успешной загрузке
// или ошибке при загрузке расшифровки текста.
func (w *Worker) downloadTranscription() {
	// Получение из базы данных списка голосовых сообщений, которые находятся на этапе "WAIT"
	// и готовы для загрузки расшифровки текста для голосового сообщения, которое было успешно обработано
	// в Salute. Для каждого такого сообщения он выполняет загрузку расшифровки текста
	// с помощью API Salute и обновляет статус в базе данных, а также отправляет пользователю сообщение
	// об успешной загрузке или ошибке при загрузке расшифровки текста.
	downloadTranscriptionData, err := w.storage.GetVoiceForDownloadTranscription()
	if err != nil {
		logrus.Error(fmt.Errorf("Worker.downloadTranscription: %w", err))
		return
	}
	// Проход по каждому голосовому сообщению, готовому для загрузки расшифровки текста,
	// и выполнение процесса загрузки расшифровки текста с помощью API Salute.
	// Если статус обработки еще не "DONE", то продолжаем ожидать, иначе обновляем статус в базе данных
	// и отправляем пользователю сообщение об успешной загрузке или ошибке при загрузке расшифровки текста.
	for i := range downloadTranscriptionData {
		// Получение расшифровки текста для голосового сообщения, которое было успешно обработано в Salute,
		// с помощью API Salute. Если статус обработки еще не "DONE", то продолжаем ожидать,
		// иначе обновляем статус в базе данных и отправляем пользователю сообщение об успешной загрузке
		// или ошибке при загрузке расшифровки текста.
		text, err := w.downloadTranscriptionExecute(downloadTranscriptionData[i])
		if err != nil {
			if errors.Is(err, fmt.Errorf("meeting processed yet")) {
				logrus.Warn(fmt.Errorf("status not DONE yet for download transcription: %d", downloadTranscriptionData[i].VoiceID))
				continue
			}
			logrus.Error(fmt.Errorf("Worker.downloadTranscription: %w", err))

			// Отправка пользователю сообщения об ошибке при загрузке расшифровки текста,
			go w.sendMsgFail(downloadTranscriptionData[i].ChatID, downloadTranscriptionData[i].VoiceID)

			// Обновление статуса обработки голосового сообщения в базе данных на "FAIL",
			if err = w.storage.SetStatusFail(downloadTranscriptionData[i].VoiceID); err != nil {
				logrus.Error(fmt.Errorf("Worker.downloadTranscription: %w", err))
			}
			continue
		}
		// Обновление статуса обработки голосового сообщения в базе данных на "WAIT"
		// и сохранение расшифровки текста для данного голосового сообщения.
		err = w.storage.SetStatusDownload(downloadTranscriptionData[i].VoiceID, text)
		if err != nil {
			logrus.Error(err)
			continue
		}
		logrus.Infof("successfull download: %d", downloadTranscriptionData[i].VoiceID)
	}
}

// createSummary - это метод, который извлекает из базы данных список голосовых сообщений,
// которые находятся на этапе "DOWNLOAD" и готовы для создания краткого содержания.
// Для каждого такого сообщения он выполняет создание краткого содержания с помощью GigaChat
// и обновляет статус обработки в базе данных, а также отправляет пользователю сообщение об успешном создании
// краткого содержания или ошибке при создании краткого содержания.
func (w *Worker) createSummary() {
	// Получение из базы данных списка голосовых сообщений, которые находятся на этапе "DOWNLOAD" и готовы для создания краткого содержания.
	// Для каждого такого сообщения он выполняет создание краткого содержания с помощью GigaChat и обновляет статус обработки в базе данных,
	// а также отправляет пользователю сообщение об успешном создании краткого содержания или ошибке при создании краткого содержания.
	createSumData, err := w.storage.GetVoiceForCreateSummary()
	if err != nil {
		logrus.Error(fmt.Errorf("Worker.createSummary: %w", err))
		return
	}
	// Проход по каждому голосовому сообщению, готовому для создания краткого содержания,
	// и выполнение процесса создания краткого содержания с помощью GigaChat.
	for i := range createSumData {
		// Получение краткого содержания для голосового сообщения с помощью GigaChat.
		// Если при создании краткого содержания произошла ошибка,
		// то отправляем пользователю сообщение об ошибке и обновляем статус в базе данных на "FAIL",
		// иначе обновляем статус на "SUCCESS"
		// и сохраняем краткое содержание для данного голосового сообщения.
		content := fmt.Sprintf(
			"сделай краткую выжимку из текста: %s\n результат должен быть информативным и без лишней воды",
			createSumData[i].Text,
		)
		summary, err := w.QuestionGigaChat(content) // w.createSummaryExecute(content)
		if err != nil {
			logrus.Error("Worker.createSummary: ", err)
			// Отправка пользователю сообщения об ошибке при создании краткого содержания,
			// указывая идентификатор голосового сообщения (voiceID) для повторной отправки.
			go w.sendMsgFail(createSumData[i].ChatID, createSumData[i].VoiceID)

			// Обновление статуса обработки голосового сообщения в базе данных на "FAIL",
			if err = w.storage.SetStatusFail(createSumData[i].VoiceID); err != nil {
				logrus.Error(fmt.Errorf("Worker.createSummary: %w", err))
			}
			continue
		}
		// Обновление статуса обработки голосового сообщения в базе данных на "SUCCESS"
		// и сохранение краткого содержания для данного голосового сообщения.
		err = w.storage.SetStatusSuccess(createSumData[i].VoiceID, summary)
		if err != nil {
			logrus.Error(err)
			continue
		}
		// Отправка пользователю сообщения об успешном создании краткого содержания,
		// указывая идентификатор голосового сообщения (voiceID) для получения краткого содержания.
		go w.sendMsgSuccess(createSumData[i].ChatID, createSumData[i].VoiceID)
		// Логирование успешного создания краткого содержания для данного голосового сообщения,
		// что может быть полезно для отслеживания процесса обработки голосовых сообщений
		// и диагностики возможных проблем в будущем.
		logrus.Infof("successfull create summary: %d", createSumData[i].VoiceID)
	}
}
