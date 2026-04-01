package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/anatolyi0311/bot-messenger/internal/models"
)

// getTokenSalute - это метод, который получает токен доступа для API Salute.
func (w *Worker) uploadExecute(upload models.UploadData) (string, error) {
	// Получение токена доступа для API Salute перед загрузкой голосового сообщения в Salute для распознавания речи, с помощью метода getTokenSalute() для дальнейшего использования при взаимодействии с API Salute.
	w.getTokenSalute()

	// Создание тела запроса для загрузки голосового сообщения в Salute для распознавания речи, используя байты голосового сообщения для создания тела запроса к API Salute.
	reqBody := bytes.NewReader(upload.VoiceData)

	// Получение идентификатора файла запроса (request_file_id) для данного голосового сообщения с помощью API Salute. Если при загрузке голосового сообщения произошла ошибка, то возвращаем ошибку, иначе возвращаем идентификатор файла запроса (request_file_id) для дальнейшей обработки.
	req, err := http.NewRequest("POST", "https://smartspeech.sber.ru/rest/v1/data:upload", reqBody)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+w.authSalute.AccessToken)
	req.Header.Set("Content-Type", "audio/ogg")

	resp, err := w.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("updload handler return status %d", resp.StatusCode)
	}

	response := models.UploadResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	return response.Result.RequestFileID, nil
}

// recognizeExecute - это метод, который выполняет распознавание речи для голосового сообщения с помощью API Salute, используя идентификатор файла запроса (request_file_id) для обработки голосового сообщения и возвращая идентификатор распознавания (recognize_id) для дальнейшей обработки.
func (w *Worker) recognizeExecute(recognize models.RecognizeData) (string, error) {
	// Получение токена доступа для API Salute перед распознаванием речи для голосового сообщения с помощью API Salute, с помощью метода getTokenSalute() для дальнейшего использования при взаимодействии с API Salute.
	w.getTokenSalute()
	requestBody := models.RecognizeRequest{
		Options: models.RecognizeRequestOptions{
			Model:         "general",
			AudioEncoding: "OPUS",
			SampleRate:    16000,
			ChannelsCount: 1,
		},
		RequestFileID: recognize.ReqFileID,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	// Получение идентификатора распознавания (recognize_id) для данного голосового сообщения с помощью API Salute. Если при распознавании речи произошла ошибка, то возвращаем ошибку, иначе возвращаем идентификатор распознавания (recognize_id) для дальнейшей обработки.
	req, err := http.NewRequest("POST", "https://smartspeech.sber.ru/rest/v1/speech:async_recognize", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+w.authSalute.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("recognize handler return status %d", resp.StatusCode)
	}

	response := models.RecognizeResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	return response.Result.ID, nil
}

// checkStatusExecute - это метод, который выполняет проверку статуса обработки голосового сообщения в Salute, используя идентификатор распознавания (recognize_id) для получения статуса обработки и возвращая идентификатор файла с результатом распознавания (resp_file_id) для дальнейшей обработки, если статус обработки "DONE", или ошибку, если статус обработки еще не "DONE" или произошла ошибка при проверке статуса.
func (w *Worker) checkStatusExecute(checkStatus models.CheckStatusData) (string, error) {
	w.getTokenSalute()

	// Получение статуса обработки голосового сообщения в Salute с помощью API Salute. Если статус обработки еще не "DONE", то возвращаем ошибку, иначе возвращаем идентификатор файла с результатом распознавания (resp_file_id) для дальнейшей обработки.
	url := fmt.Sprintf("https://smartspeech.sber.ru/rest/v1/task:get?id=%s", checkStatus.RecognizeID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+w.authSalute.AccessToken)

	resp, err := w.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("check status handler return status %d", resp.StatusCode)
	}

	response := models.ResponseCheckStatus{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}
	if response.Result.Status != "DONE" {
		return "", errors.New("status not DONE yet")
	}
	return response.Result.ResponseFileID, nil
}

// downloadTranscriptionExecute - это метод, который выполняет загрузку расшифровки текста для голосового сообщения, которое было успешно обработано в Salute, используя идентификатор файла с результатом распознавания (resp_file_id) для получения расшифровки текста и возвращая текст расшифровки для данного голосового сообщения, если запрос выполнен успешно, или ошибку, если при загрузке расшифровки текста произошла ошибка.
func (w *Worker) downloadTranscriptionExecute(download models.DownloadTranscriptionData) (string, error) {
	// Получение токена доступа для API Salute перед загрузкой расшифровки текста для голосового сообщения, которое было успешно обработано в Salute, с помощью метода getTokenSalute() для дальнейшего использования при взаимодействии с API Salute.
	w.getTokenSalute()

	// Получение расшифровки текста для голосового сообщения, которое было успешно обработано в Salute, с помощью API Salute. Если статус обработки еще не "DONE", то продолжаем ожидать, иначе возвращаем текст расшифровки для данного голосового сообщения.
	url := fmt.Sprintf("https://smartspeech.sber.ru/rest/v1/data:download?response_file_id=%s", download.RespFileID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+w.authSalute.AccessToken)

	resp, err := w.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download handler return status %d", resp.StatusCode)
	}

	response := []models.ResponseDownloadData{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	if len(response) == 0 {
		return "", fmt.Errorf("empty response from download endpoint")
	}

	// Проход по каждому элементу в ответе от API Salute и объединение текста расшифровки для данного голосового сообщения, возвращая текст расшифровки для данного голосового сообщения, если запрос выполнен успешно, или ошибку, если при загрузке расшифровки текста произошла ошибка.
	var text string
	for i := range response {
		for j := range response[i].Results {
			if text != "" {
				text += " "
			}
			text += response[i].Results[j].Text
		}
	}

	return text, nil
}

// createSummaryExecute - это метод, который выполняет создание краткого содержания для текста распознанной речи с помощью GigaChat, используя текст распознанной речи для создания запроса к API GigaChat и возвращая краткое содержание для данного текста распознанной речи, если запрос выполнен успешно, или ошибку, если при создании краткого содержания произошла ошибка.
func (w *Worker) createSummaryExecute(summaryData models.CreateSummaryData) (string, error) {
	content := fmt.Sprintf("сделай краткую выжимку из текста: %s\n результат должен быть информативным и без лишней воды", summaryData.Text)
	requestBody := models.GigaChatRequest{
		Model:             "GigaChat",
		Stream:            false,
		RepetitionPenalty: 1,
		Messages: []models.Message{
			{
				Role:    "user",
				Content: content,
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	// Получение краткого содержания для текста распознанной речи с помощью API GigaChat. Если при создании краткого содержания произошла ошибка, то возвращаем ошибку, иначе возвращаем краткое содержание для данного текста распознанной речи.
	req, err := http.NewRequest("POST", "https://gigachat.devices.sberbank.ru/api/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+w.authGigaChat.AccessToken)

	resp, err := w.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("create summary handler return status %d", resp.StatusCode)
	}

	response := models.GigaChatResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}
	if len(response.Choices) < 1 {
		return "", fmt.Errorf("create summary handler return status empty response")
	}
	var summary string
	for i := range response.Choices {
		summary += response.Choices[i].Message.Content
	}

	return summary, nil
}
