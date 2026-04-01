package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func (w *Worker) getTokenSalute() {
	// Если токен для Salute еще действителен, то не нужно запрашивать новый.
	if w.authSalute.ExpiresAt.After(time.Now()) {
		return
	}

	// Генерация уникального идентификатора запроса (RqUID) для отслеживания запросов к API Salute.
	rquid := uuid.New().String()
	// Формирование тела запроса для получения токена доступа к API Salute.
	// В данном случае, указывается область доступа (scope) для получения прав на использование функций Salute.
	body := fmt.Sprintf("scope=%s", "SALUTE_SPEECH_PERS")
	reqBody := strings.NewReader(body)

	// Создание HTTP-запроса для получения токена доступа к API Salute. Запрос отправляется на эндпоинт OAuth сервера Salute.
	req, err := http.NewRequest("POST", "https://ngw.devices.sberbank.ru:9443/api/v2/oauth", reqBody)
	if err != nil {
		logrus.Error("Worker.getTokenSalute(): ", err)
		return
	}

	// Формирование заголовков запроса, включая Content-Type, RqUID и Authorization.
	// Заголовок Authorization содержит базовую аутентификацию, которая кодирует ClientID и ClientSecret
	// в формате Base64 для обеспечения безопасности при передаче учетных данных.
	authKey := base64.StdEncoding.EncodeToString([]byte(w.cfg.Salute.ClientID + ":" + w.cfg.Salute.ClientSecret))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("RqUID", rquid)
	req.Header.Set("Authorization", "Basic "+authKey)

	// Отправка HTTP-запроса и обработка ответа от сервера Salute.
	// Если запрос выполнен успешно, то из ответа извлекается токен доступа и время его истечения,
	// которые сохраняются в структуре Auth для дальнейшего использования при взаимодействии с API Salute.
	resp, err := w.client.Do(req)
	if err != nil {
		logrus.Error("Worker.getTokenSalute(): ", err)
		return
	}
	defer resp.Body.Close()

	// Проверка статуса ответа от сервера Salute. Если статус не OK, то логируется ошибка и функция завершается без сохранения токена.
	if resp.StatusCode != http.StatusOK {
		logrus.Error("Worker.getTokenSalute(): ", resp.StatusCode)
		return
	}

	// Чтение тела ответа и извлечение токена доступа и времени его истечения из JSON-ответа от сервера Salute.
	respBody, _ := io.ReadAll(resp.Body)

	// Извлечение токена доступа и времени его истечения из JSON-ответа от сервера Salute
	var result map[string]interface{}
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		logrus.Error("Worker.getTokenSalute(): ", err)
	}

	// сохранение их в структуре Auth для дальнейшего использования при взаимодействии с API Salute.
	token := result["access_token"]
	expiresAt := result["expires_at"]

	// Сохранение токена доступа и времени его истечения в структуре Auth для дальнейшего использования при взаимодействии с API Salute.
	w.authSalute.AccessToken = token.(string)
	w.authSalute.ExpiresAt = time.Unix(int64(expiresAt.(float64)), 0)

	logrus.Info("successfully obtained token, expires at ", w.authSalute.ExpiresAt)
}

// getTokenGigaChat - это метод, который получает токен доступа для API GigaChat.
// Он проверяет, действителен ли текущий токен, и если нет, то отправляет запрос на получение нового токена,
// используя базовую аутентификацию с ClientID и ClientSecret.
// Полученный токен и время его истечения сохраняются в структуре Auth для дальнейшего использования при взаимодействии с API GigaChat.
func (w *Worker) getTokenGigaChat() {
	if w.authGigaChat.ExpiresAt.After(time.Now()) {
		return
	}

	// Генерация уникального идентификатора запроса (RqUID) для отслеживания запросов к API GigaChat.
	rquid := uuid.New().String()
	authKey := base64.StdEncoding.EncodeToString([]byte(w.cfg.GigaChat.ClientID + ":" + w.cfg.GigaChat.ClientSecret))
	body := []byte("scope=GIGACHAT_API_PERS")

	// Создание HTTP-запроса для получения токена доступа к API GigaChat. Запрос отправляется на эндпоинт OAuth сервера GigaChat.
	req, err := http.NewRequest("POST", "https://ngw.devices.sberbank.ru:9443/api/v2/oauth", bytes.NewBuffer(body))
	if err != nil {
		logrus.Error("Worker.getTokenGigaChat(): ", err)
		return
	}

	// Формирование заголовков запроса, включая Content-Type, Accept, RqUID и Authorization.
	// Заголовок Authorization содержит базовую аутентификацию, которая кодирует ClientID и ClientSecret
	// в формате Base64 для обеспечения безопасности при передаче учетных данных.
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("RqUID", rquid)
	req.Header.Set("Authorization", "Basic "+authKey)

	// Отправка HTTP-запроса и обработка ответа от сервера GigaChat.
	// Если запрос выполнен успешно, то из ответа извлекается токен доступа и время его истечения,
	// которые сохраняются в структуре Auth для дальнейшего использования при взаимодействии с API GigaChat.
	resp, err := w.client.Do(req)
	if err != nil {
		logrus.Error("Worker.getTokenGigaChat(): ", err)
		return
	}
	defer resp.Body.Close()

	// Проверка статуса ответа от сервера GigaChat. Если статус не OK, то логируется ошибка и функция завершается без сохранения токена.
	if resp.StatusCode != http.StatusOK {
		logrus.Error("Worker.getTokenGigaChat(): ", resp.StatusCode)
		return
	}

	// Чтение тела ответа и извлечение токена доступа и времени его истечения из JSON-ответа от сервера GigaChat.
	respBody, _ := io.ReadAll(resp.Body)

	// Извлечение токена доступа и времени его истечения из JSON-ответа от сервера GigaChat
	var result map[string]interface{}
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		logrus.Error("Worker.getTokenGigaChat(): ", err)
	}

	// сохранение их в структуре Auth для дальнейшего использования при взаимодействии с API GigaChat.
	token := result["access_token"]
	expiresAt := result["expires_at"]

	// Сохранение токена доступа и времени его истечения в структуре Auth для дальнейшего использования при взаимодействии с API GigaChat.
	w.authGigaChat.AccessToken = token.(string)
	w.authGigaChat.ExpiresAt = time.Unix(int64(expiresAt.(float64)), 0)

	logrus.Info("successfully obtained token, expires at ", w.authGigaChat.ExpiresAt)
}
