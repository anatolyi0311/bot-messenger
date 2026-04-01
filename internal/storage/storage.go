// Package storage - это пакет, который содержит логику взаимодействия с базой данных,
package storage

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/anatolyi0311/bot-messenger/internal/models"
)

// Storage - это структура, которая будет использоваться для взаимодействия с базой данных.
type Storage struct {
	db *sql.DB
}

// New - это функция, которая инициализирует новый экземпляр Storage с заданной базой данных.
func New(db *sql.DB) *Storage {
	return &Storage{db: db}
}

// SaveIncomingVoice - это метод, который сохраняет входящее голосовое сообщение в базе данных и возвращает его идентификатор (voiceID) для дальнейшей обработки.
func (s *Storage) SaveIncomingVoice(voiceBytes []byte, chatID int64) (int, error) {
	var voiceID int
	err := s.db.QueryRow(`
		INSERT INTO voice_recognize (voice_data, chat_id, process_step) 
        VALUES ($1, $2, $3)
		RETURNING id
	`, voiceBytes, chatID, "BEGIN").Scan(&voiceID)
	return voiceID, err
}

// GetVoiceForUpload - это метод, который извлекает из базы данных список голосовых сообщений, которые находятся на этапе "BEGIN" и готовы для загрузки в Salute для распознавания речи.
func (s *Storage) GetVoiceForUpload() ([]models.UploadData, error) {
	rows, err := s.db.Query(`
		SELECT id, chat_id, voice_data
		FROM voice_recognize
		WHERE process_step = $1
	`, "BEGIN")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var uploadDataList []models.UploadData

	for rows.Next() {
		var uploadData models.UploadData
		err := rows.Scan(&uploadData.VoiceID, &uploadData.ChatID, &uploadData.VoiceData)
		if err != nil {
			return nil, err
		}
		uploadDataList = append(uploadDataList, uploadData)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	if len(uploadDataList) < 1 {
		return nil, errors.New("no voices for upload")
	}
	return uploadDataList, nil
}

// SetStatusFail - это метод, который обновляет статус обработки голосового сообщения в базе данных на "FAIL" для заданного идентификатора голосового сообщения (voiceID).
func (s *Storage) SetStatusFail(voiceID int64) error {
	result, err := s.db.Exec(`
		UPDATE voice_recognize
		SET process_step = $2
		WHERE id = $1
	`, voiceID, "FAIL")

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("storage.SetStatusFail(): rows affected = 0 for voice: %d", voiceID)
	}
	return nil
}

// SetStatusUpload - это метод, который обновляет статус обработки голосового сообщения в базе данных на "UPLOAD" и сохраняет идентификатор файла запроса (reqFileID) для заданного идентификатора голосового сообщения (voiceID).
func (s *Storage) SetStatusUpload(voiceID int64, reqFileID string) error {
	_, err := s.db.Exec(`
		UPDATE voice_recognize 
		SET process_step = $2, request_file_id = $3
		WHERE id = $1
	`, voiceID, "UPLOAD", reqFileID)
	return err
}
