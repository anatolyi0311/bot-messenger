// Package service - это пакет, который содержит бизнес-логику для обработки голосовых сообщений,
// взаимодействия с базой данных и Salute, 
// а также предоставления методов для сохранения входящих голосовых сообщений и
// получения краткого содержания по идентификатору встречи.
package service

import (
	"strconv"
	"strings"
)

// SaveIncomingID - это метод, который сохраняет идентификатор чата (chatID) в базе данных 
// и возвращает его идентификатор (voiceID) для дальнейшей обработки.
func (w *Worker) SaveIncomingID(chatID int64) (int, error) {
	return w.storage.SaveIncomingID(chatID)
}

// SaveIncomingVoice - это метод, который сохраняет входящее голосовое сообщение в базе данных 
// и возвращает его идентификатор (voiceID) для дальнейшей обработки.
func (w *Worker) SaveIncomingVoice(voiceBytes []byte, formatAudio string, chatID int64) (int, error) {
	return w.storage.SaveIncomingVoice(voiceBytes, formatAudio, chatID)
}

// GetSummaryByID - это метод, который получает краткое содержание по идентификатору встречи (voiceID) 
// и идентификатору чата (chatID) из базы данных для предоставления пользователю.
func (w *Worker) GetSummaryByID(voiceID, chatID int64) (string, error) {
	return w.storage.GetSummaryByID(voiceID, chatID)
}

// GetListSummaryID - это метод, который получает список идентификаторов встреч (IDs) 
// по идентификатору чата (chatID) из базы данных для предоставления пользователю.
func (w *Worker) GetListSummaryID(chatID int64) ([]int64, error) {
	return w.storage.GetListSummaryID(chatID)
}

// ListResponseBuilder - это метод, который принимает срез идентификаторов встреч (IDs) 
// и строит строку ответа, которая содержит список этих встреч для предоставления пользователю.
func (w *Worker) ListResponseBuilder(IDs []int64) string {
	var builder strings.Builder
	builder.WriteString("Список ваших встреч:\n")
	for i := range IDs {
		builder.WriteString("Встреча ")
		builder.WriteString(strconv.FormatInt(IDs[i], 10))
		builder.WriteString("\n")
	}
	return builder.String()
}

// FindByKey - это метод, который принимает идентификатор чата (chatID) и срез ключевых слов (keyWords), 
// преобразует каждое ключевое слово в шаблон для поиска в базе данных 
// и возвращает список идентификаторов встреч, которые соответствуют этим шаблонам.
func (w *Worker) FindByKey(chatID int64, keyWords []string) ([]int64, error) {
	patterns := make([]string, len(keyWords))
	for i := range keyWords {
		var builder strings.Builder
		builder.WriteString("%")
		builder.WriteString(keyWords[i])
		builder.WriteString("%")
		patterns[i] = builder.String()
	}
	return w.storage.FindByKey(chatID, patterns)
}

// GigaChatReqBuilder - это метод, который строит строку запроса для GigaChat, 
// объединяя все элементы из переданного среза строк (args) в одну строку с пробелами между ними.
func (w *Worker) GigaChatReqBuilder(args []string) string {
	var builder strings.Builder
	for i := range args {
		builder.WriteString(args[i])
		builder.WriteString(" ")
	}
	return builder.String()
}
