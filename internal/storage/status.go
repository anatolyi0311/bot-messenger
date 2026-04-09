// Package storage - это пакет, который содержит методы для обновления статуса обработки голосовых сообщений в базе данных.
package storage

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// SetStatusRecognition - это метод, который обновляет статус обработки голосового сообщения
// в базе данных на "RECOGNITION" и сохраняет идентификатор распознавания (recognizeID)
// для заданного идентификатора голосового сообщения (voiceID).
func (s *Storage) SetStatusRecognition(voiceID int64, recognizeID string) error {
	_, err := s.db.Exec(`
		UPDATE voice_recognize 
		SET process_step = $2, recognize_id = $3
		WHERE id = $1
	`, voiceID, "RECOGNITION", recognizeID)
	return err
}

// SetStatusFail - это метод, который обновляет статус обработки голосового сообщения
// в базе данных на "FAIL" для заданного идентификатора голосового сообщения (voiceID).
func (s *Storage) SetStatusFail(voiceID int64) error {
	// Обновляем статус обработки голосового сообщения на "FAIL"
	result, err := s.db.Exec(`
		UPDATE voice_recognize
		SET process_step = $2
		WHERE id = $1
	`, voiceID, "FAIL")
	// Логируем результат обновления статуса для отладки и диагностики ошибок
	if err != nil {
		logrus.Error("Storage.SetStatusFail-1: ", err)
		return err
	}
	// Проверяем количество затронутых строк, чтобы убедиться, что статус был обновлен
	// для существующей записи в базе данных и логируем результат для отладки и диагностики ошибок (например,
	// если rowsAffected равно 0, это может указывать на то, что запись с данным voiceID
	// не существует в базе данных). Если rowsAffected равно 0, то возвращаем ошибку,
	// которая может быть обработана вызывающим кодом для информирования пользователя о том,
	// что произошла ошибка при обновлении статуса. Если rowsAffected больше 0,
	// то статус был успешно обновлен и мы возвращаем nil, что указывает на успешное выполнение метода.
	// Если rowsAffected равно 0, то это может указывать на то, что запись с данным voiceID
	// не существует в базе данных, и мы логируем эту информацию для отладки и диагностики ошибок.
	// Если rowsAffected больше 0, то статус был успешно обновлен и мы возвращаем nil, что указывает на успешное выполнение метода.
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logrus.Error("Storage.SetStatusFail-2: ", err)
		return err
	}
	// Логируем количество затронутых строк для отладки и диагностики ошибок
	logrus.Infof("Storage.SetStatusFail-3: rows affected = %d for voiceID = %d", rowsAffected, voiceID)
	if rowsAffected == 0 {
		logrus.Error(fmt.Errorf("Storage.SetStatusFail-3: rows affected = 0 for voice: %d", voiceID))
		return err // fmt.Errorf("Storage.SetStatusFail: rows affected = 0 for voice: %d", voiceID)
	}
	// Если rowsAffected больше 0, то статус был успешно обновлен и мы возвращаем nil,
	// что указывает на успешное выполнение метода.
	return nil
}

// SetStatusUpload - это метод, который обновляет статус обработки голосового сообщения
// в базе данных на "UPLOAD" и сохраняет идентификатор файла запроса (reqFileID)
// для заданного идентификатора голосового сообщения (voiceID).
func (s *Storage) SetStatusUpload(voiceID int64, reqFileID string) error {
	// Обновляем статус обработки голосового сообщения на "UPLOAD" и сохраняем идентификатор файла запроса
	_, err := s.db.Exec(`
		UPDATE voice_recognize 
		SET process_step = $2, request_file_id = $3
		WHERE id = $1
	`, voiceID, "UPLOAD", reqFileID)
	// Логируем результат обновления статуса для отладки и диагностики ошибок (например,
	// если произошла ошибка при выполнении запроса к базе данных, то мы логируем эту ошибку
	// для дальнейшего анализа и устранения проблемы).
	// Если обновление статуса прошло успешно, то мы логируем информацию о том,
	// что статус был обновлен на "UPLOAD" для данного voiceID и reqFileID,
	// что может быть полезно для отслеживания процесса обработки голосовых сообщений и диагностики возможных проблем в будущем.
	// if err != nil {
	// 	logrus.Error("Storage.SetStatusUpload: ", err)
	// }
	// logrus.Infof("Storage.SetStatusUpload: status updated to 'UPLOAD' for voiceID = %d, reqFileID = %s", voiceID, reqFileID)
	logrus.Infof("SetStatusUpload: voiceID = %d, reqFileID = %s", voiceID, reqFileID)
	return err
}

// SetStatusWait - это метод, который обновляет статус обработки голосового сообщения
// в базе данных на "WAIT" и сохраняет идентификатор файла ответа (respFileID)
// для заданного идентификатора голосового сообщения (voiceID).
func (s *Storage) SetStatusWait(voiceID int64, respFileID string) error {
	// Обновляем статус обработки голосового сообщения на "WAIT" и сохраняем идентификатор файла ответа
	_, err := s.db.Exec(`
		UPDATE voice_recognize 
		SET process_step = $2, response_file_id = $3
		WHERE id = $1
	`, voiceID, "WAIT", respFileID)
	// Логируем результат обновления статуса для отладки и диагностики ошибок (например,
	// если произошла ошибка при выполнении запроса к базе данных, то мы логируем эту ошибку
	// для дальнейшего анализа и устранения проблемы).
	// Если обновление статуса прошло успешно, то мы логируем информацию о том,
	// что статус был обновлен на "WAIT" для данного voiceID и respFileID,
	// что может быть полезно для отслеживания процесса обработки голосовых сообщений и диагностики возможных проблем в будущем.
	// if err != nil {
	// 	logrus.Error("Storage.SetStatusWait: ", err)
	// }
	// logrus.Infof("Storage.SetStatusWait: status updated to 'WAIT' for voiceID = %d, respFileID = %s", voiceID, respFileID)
	logrus.Infof("SetStatusWait: voiceID = %d, respFileID = %s", voiceID, respFileID)
	return err
}

// SetStatusDownload - это метод, который обновляет статус обработки голосового сообщения
// в базе данных на "DOWNLOAD" и сохраняет распознанный текст (text)
// для заданного идентификатора голосового сообщения (voiceID).
func (s *Storage) SetStatusDownload(voiceID int64, text string) error {
	// Обновляем статус обработки голосового сообщения на "DOWNLOAD" и сохраняем распознанный текст
	_, err := s.db.Exec(`
		UPDATE voice_recognize 
		SET process_step = $2, transcription = $3
		WHERE id = $1
	`, voiceID, "DOWNLOAD", text)
	logrus.Infof("SetStatusDownload: voiceID = %d, text = %s", voiceID, text)
	return err
}

// SetStatusSuccess - это метод, который обновляет статус обработки голосового сообщения
// в базе данных на "SUCCESS", сохраняет краткое содержание (summary) и время создания
// для заданного идентификатора голосового сообщения (voiceID).
func (s *Storage) SetStatusSuccess(voiceID int64, summary string) error {
	// Обновляем статус обработки голосового сообщения на "SUCCESS", сохраняем краткое содержание и время создания
	_, err := s.db.Exec(`
		UPDATE voice_recognize
		SET process_step = $2, summary = $3, created_at = $4
		WHERE id = $1
	`, voiceID, "SUCCESS", summary, time.Now())
	// Логируем результат обновления статуса для отладки и диагностики ошибок (например,
	// если произошла ошибка при выполнении запроса к базе данных, то мы логируем эту ошибку
	// для дальнейшего анализа и устранения проблемы).
	// Если обновление статуса прошло успешно, то мы логируем информацию о том,
	// что статус был обновлен на "SUCCESS" для данного voiceID и краткого содержания,
	// что может быть полезно для отслеживания процесса обработки голосовых сообщений и диагностики возможных проблем в будущем.
	// if err != nil {
	// 	logrus.Error("Storage.SetStatusSuccess: ", err)
	// }
	// logrus.Infof("Storage.SetStatusSuccess: status updated to 'SUCCESS' for voiceID = %d, summary = %s", voiceID, summary)
	logrus.Infof("SetStatusSuccess: voiceID = %d, summary = %s", voiceID, summary)
	return err
}
