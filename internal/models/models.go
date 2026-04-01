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
type RecognizeData struct {
	VoiceID   int64
	ChatID    int64
	ReqFileID string
}

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
