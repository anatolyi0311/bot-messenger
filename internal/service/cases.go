// Package service - это пакет, который содержит бизнес-логику для обработки голосовых сообщений,
// взаимодействия с базой данных и Salute, а также предоставления методов для сохранения входящих голосовых сообщений и
// получения краткого содержания по идентификатору встречи.
package service

// SaveIncomingID - это метод, который сохраняет идентификатор чата (chatID) в базе данных и возвращает его идентификатор (voiceID) для дальнейшей обработки.
func (w *Worker) SaveIncomingID(chatID int64) (int, error) {
	return w.storage.SaveIncomingID(chatID)
}

// SaveIncomingVoice - это метод, который сохраняет входящее голосовое сообщение в базе данных и возвращает его идентификатор (voiceID) для дальнейшей обработки.
func (w *Worker) SaveIncomingVoice(voiceBytes []byte, chatID int64) (int, error) {
	return w.storage.SaveIncomingVoice(voiceBytes, chatID)
}

// GetSummaryByID - это метод, который получает краткое содержание по идентификатору встречи (voiceID) и идентификатору чата (chatID) из базы данных для предоставления пользователю.
func (w *Worker) GetSummaryByID(voiceID, chatID int64) (string, error) {
	return w.storage.GetSummaryByID(voiceID, chatID)
}
