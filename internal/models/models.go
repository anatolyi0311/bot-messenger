// Package models содержит определения структур данных, используемых в приложении.
package models

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
