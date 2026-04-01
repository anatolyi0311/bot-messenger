// Package models содержит определения структур данных, используемых в приложении.
package models

// MsgErrSaveVoice - это константа, которая содержит сообщение об ошибке при сохранении голосового сообщения в базе данных.
const MsgErrSaveVoice = "Error saving voice message. Please try again later."

// MsgSuccessSaveVoice - это константа, которая содержит сообщение об успешном сохранении голосового сообщения в базе данных, включая идентификатор голосового сообщения (voiceID) для дальнейшей обработки.
const MsgSuccessSaveVoice = "Voice message saved successfully with ID: %d."

// UserMessage - это структура, которая представляет сообщение от пользователя,
// содержащее идентификатор чата и текст сообщения.
type UserMessage struct {
	ChatID  int64
	Message string
}

// UploadData - это структура, которая представляет данные для загрузки,
type UploadData struct {
	VoiceID   int64
	ChatID    int64
	VoiceData []byte
}

// resultUpload - это структура, которая представляет результат загрузки голосового сообщения в Salute, содержащая идентификатор файла запроса (request_file_id) для дальнейшей обработки.
type resultUpload struct {
	RequestFileID string `json:"request_file_id"`
}

// UploadResponse - это структура, которая представляет ответ от сервера Salute при загрузке голосового сообщения, содержащая результат загрузки, который включает идентификатор файла запроса (request_file_id) для дальнейшей обработки.
type UploadResponse struct {
	Result resultUpload `json:"result"`
}

// RecognizeRequest - это структура, которая представляет запрос на распознавание речи в Salute, содержащая параметры для распознавания и идентификатор файла запроса (request_file_id) для обработки голосового сообщения.
type RecognizeRequest struct {
	Options       RecognizeRequestOptions `json:"options"`
	RequestFileID string                  `json:"request_file_id"`
}

// RecognizeData - это структура, которая представляет данные для распознавания речи, содержащая идентификатор голосового сообщения (voiceID), идентификатор чата (chatID) и идентификатор файла запроса (reqFileID) для обработки голосового сообщения с помощью Salute.
type RecognizeData struct {
	VoiceID   int64
	ChatID    int64
	ReqFileID string
}

// RecognizeRequestOptions - это структура, которая представляет параметры для распознавания речи в Salute, включая модель распознавания, кодировку аудио, частоту дискретизации и количество каналов для обработки голосового сообщения.
type RecognizeRequestOptions struct {
	Model         string `json:"model"`
	AudioEncoding string `json:"audio_encoding"`
	SampleRate    int    `json:"sample_rate"`
	ChannelsCount int    `json:"channels_count"`
}

// RecognizeResponse - это структура, которая представляет ответ от сервера Salute при распознавании речи, содержащая результат распознавания, который включает идентификатор распознавания (id) для дальнейшей обработки.
type RecognizeResponse struct {
	Result resultRecognize `json:"result"`
}

type resultRecognize struct {
	ID string `json:"id"`
}

// GetTokenSaluteResponse - это структура, которая представляет ответ от сервера Salute при получении токена доступа, содержащая токен доступа (access_token) для дальнейшего использования при взаимодействии с API Salute.
type GetTokenSaluteResponse struct {
	AccessToken string `json:"access_token"`
}

// GetTokenGigaChatResponse - это структура, которая представляет ответ от сервера GigaChat при получении токена доступа, содержащая токен доступа (access_token) для дальнейшего использования при взаимодействии с API GigaChat.
type GetTokenGigaChatResponse struct {
	AccessToken string `json:"access_token"`
}

// RecognizeResultResponse - это структура, которая представляет ответ от сервера Salute при получении результата распознавания, содержащая текст распознанной речи (text) для дальнейшего использования.
type RecognizeResultResponse struct {
	Result resultRecognizeResult `json:"result"`
}

type resultRecognizeResult struct {
	Text string `json:"text"`
}

// ResponseGigaChat - это структура, которая представляет ответ от сервера GigaChat при отправке сообщения, содержащая результат отправки, который включает идентификатор сообщения (message_id) для дальнейшей обработки.
type ResponseGigaChat struct {
	Result resultGigaChat `json:"result"`
}

type resultGigaChat struct {
	MessageID int64 `json:"message_id"`
}

// ResponseCheckStatus - это структура, которая представляет ответ от сервера Salute при проверке статуса распознавания речи, содержащая статус распознавания (status) и идентификатор файла ответа (response_file_id) для дальнейшей обработки.
type ResponseCheckStatus struct {
	Result resultCheckStatus `json:"result"`
}

type resultCheckStatus struct {
	Status         string `json:"status"`
	ResponseFileID string `json:"response_file_id"`
}

// GetTokenGigaChatResponse - это структура, которая представляет ответ от сервера GigaChat при получении токена доступа, содержащая токен доступа (access_token) для дальнейшего использования при взаимодействии с API GigaChat.
type ResponseFileID struct {
	Result resultFileID `json:"result"`
}

type resultFileID struct {
	FileID string `json:"file_id"`
}

// CheckStatusData - это структура, которая представляет данные для проверки статуса распознавания речи, содержащая идентификатор голосового сообщения (voiceID), идентификатор чата (chatID) и идентификатор распознавания (recognizeID) для проверки статуса распознавания речи с помощью Salute.
type CheckStatusData struct {
	VoiceID     int64
	ChatID      int64
	RecognizeID string
}

// RecognizeResultData - это структура, которая представляет данные для получения результата распознавания речи, содержащая идентификатор голосового сообщения (voiceID), идентификатор чата (chatID) и идентификатор распознавания (recognizeID) для получения результата распознавания речи с помощью Salute.
type RecognizeResultData struct {
	VoiceID     int64
	ChatID      int64
	RecognizeID string
}
