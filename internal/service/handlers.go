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
	w.getTokenSalute()

	reqBody := bytes.NewReader(upload.VoiceData)

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

func (w *Worker) recognizeExecute(recognize models.RecognizeData) (string, error) {
	// must return recognize_id
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

func (w *Worker) checkStatusExecute(checkStatus models.CheckStatusData) (string, error) {
	w.getTokenSalute()

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
